package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
)

type Dispatcher interface {
	FindTenantIDByDomainName(ctx context.Context, domainName string) (string, error)
}

func NewTenantResolution(tenantConv Dispatcher) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// A. ホスト名の取得
			// Nginxが付与した X-Forwarded-Host を優先し、なければ Host を使用
			host := r.Header.Get("X-Forwarded-Host")
			if host == "" {
				host = r.Host
			}

			// ポート番号が含まれる場合は除去 (例: localhost:8888 -> localhost)
			if strings.Contains(host, ":") {
				host = strings.Split(host, ":")[0]
			}

			// B. テナントの特定ロジック
			// 例: tenant-a.example.com -> "tenant-a" を抽出
			parts := strings.Split(host, ".")
			tenantName := parts[0]

			// ローカル開発用などでIP直打ちの場合のフォールバックなどを適宜追加
			// if tenantName == "localhost" || tenantName == "127" {
			// 	tenantName = "default-tenant"
			// }

			slog.Debug("Resolved tenant name: %s from host: %s", tenantName, host)

			// C. Contextにテナント情報を保存
			tenantID, err := tenantConv.FindTenantIDByDomainName(r.Context(), tenantName)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			ctx := context.WithValue(r.Context(), KeyTenantDomainName, tenantName)
			ctx = context.WithValue(ctx, KeyTenantID, tenantID)

			// 次の処理へContextを引き継いでリクエストを回す
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
