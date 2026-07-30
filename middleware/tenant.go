package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
)

// Dispatcher はテナント名からテナントIDを解決するためのインターフェースです。
// NewTenantResolution がリクエストごとに呼び出します。
type Dispatcher interface {
	// FindTenantIDByDomainName はテナント名（リクエストホストの最初の DNS ラベル。
	// 例: "tenant-a.example.com" の "tenant-a"）からテナントIDを解決します。
	// エラーを返した場合、該当リクエストは HTTP 400 で終了します。
	FindTenantIDByDomainName(ctx context.Context, domainName string) (string, error)
}

// テナント解決をスキップするパス
var tenantSkipPaths = map[string]bool{
	"/openapi.json": true,
	"/openapi.yaml": true,
	"/docs":         true,
	"/health":       true,
}

// parseTrustedProxyCIDRs は環境変数 TRUSTED_PROXY_CIDRS (カンマ区切りの CIDR) を解析します。
// 例: "10.0.0.0/8,192.168.0.0/16"
func parseTrustedProxyCIDRs() []*net.IPNet {
	raw := strings.TrimSpace(os.Getenv("TRUSTED_PROXY_CIDRS"))
	if raw == "" {
		return nil
	}
	var nets []*net.IPNet
	for _, c := range strings.Split(raw, ",") {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if _, n, err := net.ParseCIDR(c); err == nil {
			nets = append(nets, n)
		} else {
			slog.Warn("invalid TRUSTED_PROXY_CIDRS entry ignored", slog.String("cidr", c), slog.String("error", err.Error()))
		}
	}
	return nets
}

// isTrustedProxy はリクエストの直接の送信元 (RemoteAddr) が信頼できるプロキシかを判定します。
//
// 信頼するのは環境変数 TRUSTED_PROXY_CIDRS で明示的に指定された範囲のみです（fail-closed）。
// 未設定の場合はいかなる送信元も信頼しないため、X-Forwarded-Host / X-Forwarded-For /
// X-Real-IP といったクライアント制御のヘッダーは採用されません。
//
// プライベートアドレスやループバックを暗黙に信頼すると、Kubernetes や共有 VPC のように
// 同一プライベート網へ第三者のワークロードが同居する環境で、テナント成りすまし
// (tenant.go) やレート制限の回避 (ratelimit.go) を許してしまうため、明示指定を必須とします。
// リバースプロキシの背後で運用する場合は、そのプロキシのアドレス範囲を
// TRUSTED_PROXY_CIDRS に設定してください（例: "10.0.0.0/8,127.0.0.1/32"）。
func isTrustedProxy(remoteAddr string, trusted []*net.IPNet) bool {
	if len(trusted) == 0 {
		return false
	}
	host := remoteAddr
	if h, _, err := net.SplitHostPort(remoteAddr); err == nil {
		host = h
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, n := range trusted {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// NewTenantResolution はリクエストのホスト名からテナントを解決するミドルウェアを返します。
// /openapi.json, /openapi.yaml, /docs, /health は解決をスキップします。
// ホスト（ポート除去済み）の最初の DNS ラベルをテナント名として抽出します。
// X-Forwarded-Host は TRUSTED_PROXY_CIDRS で明示的に信頼したプロキシ経由の場合のみ
// 採用します。解決に失敗した場合は 400 を返して処理を中断します。
// 成功時は KeyTenantDomainName / KeyTenantID / KeyTenantHost を Context に格納します。
// パターン: Chi (r.Use で適用)
func NewTenantResolution(tenantConv Dispatcher) func(http.Handler) http.Handler {
	trustedProxies := parseTrustedProxyCIDRs()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// テナント解決が不要なシステムパスはスキップ
			if tenantSkipPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// A. ホスト名の取得
			// X-Forwarded-Host は TRUSTED_PROXY_CIDRS で明示的に信頼したプロキシを
			// 経由したリクエストの場合のみ採用する。未設定時は採用しない。
			// 直接公開された経路でクライアントが X-Forwarded-Host を偽装して
			// 任意テナントへ成りすますことを防ぐ。
			host := r.Host
			if isTrustedProxy(r.RemoteAddr, trustedProxies) {
				if fwd := r.Header.Get("X-Forwarded-Host"); fwd != "" {
					host = fwd
				}
			}

			// ポート番号が含まれる場合は除去 (例: localhost:8888 -> localhost)
			if strings.Contains(host, ":") {
				host = strings.Split(host, ":")[0]
			}

			// B. テナントの特定ロジック
			// 例: tenant-a.example.com -> "tenant-a" を抽出
			parts := strings.Split(host, ".")
			tenantName := parts[0]

			slog.Debug("Resolved tenant", slog.String("tenant_name", tenantName), slog.String("host", host))

			// C. Contextにテナント情報を保存
			tenantID, err := tenantConv.FindTenantIDByDomainName(r.Context(), tenantName)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			ctx := context.WithValue(r.Context(), KeyTenantDomainName, tenantName)
			ctx = context.WithValue(ctx, KeyTenantID, tenantID)
			// 検証済みホスト（ポート除去済み）を保存する。Origin を信用しない URL 生成に使用する。
			ctx = context.WithValue(ctx, KeyTenantHost, host)

			// 次の処理へContextを引き継いでリクエストを回す
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
