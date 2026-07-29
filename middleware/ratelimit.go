package middleware

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// rateLimitClientIP はレート制限のキーに用いるクライアント識別子（IP）を返します。
// X-Real-IP / X-Forwarded-For は信頼できるプロキシ（TRUSTED_PROXY_CIDRS、未設定時は
// プライベート/ループバック）経由のリクエストの場合のみ採用し、それ以外は直接の
// 接続元アドレスを用います。これにより、直接公開された経路でクライアントがこれらの
// ヘッダーを偽装し、毎回異なるキーでレート制限を回避することを防ぎます。
func rateLimitClientIP(ctx huma.Context, trusted []*net.IPNet) string {
	remote := ctx.RemoteAddr()
	if isTrustedProxy(remote, trusted) {
		if xrip := strings.TrimSpace(ctx.Header("X-Real-IP")); xrip != "" {
			return xrip
		}
		if xff := ctx.Header("X-Forwarded-For"); xff != "" {
			if ip := strings.TrimSpace(strings.Split(xff, ",")[0]); ip != "" {
				return ip
			}
		}
	}
	// 信頼できないピア → 直接の接続元アドレスで識別する
	host := remote
	if h, _, err := net.SplitHostPort(remote); err == nil {
		host = h
	}
	return host
}

// RateLimitConfig レート制限の設定構造体
type RateLimitConfig struct {
	Rate   int           // 期間あたりの許可リクエスト数
	Burst  int           // バースト（瞬間的な許容超過数）
	Period time.Duration // 期間 (例: 1秒, 1分)
	Name   string        // レートリミット識別子 (例: "global", "login")
}

// DefaultRateLimitConfig はデフォルトのレート制限設定を返します。
// 1秒あたり10リクエスト、バースト20を許容します。
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Rate:   10,
		Burst:  20,
		Period: 1 * time.Second,
		Name:   "default",
	}
}

// NewRedisRateLimiter は Redis ベースのレート制限 Huma ミドルウェアを生成します。
// パターン: Huma (api.UseMiddleware で適用)
func NewRedisRateLimiter(rdb *redis.Client, config RateLimitConfig) func(huma.Context, func(huma.Context)) {
	// redis_rate ライブラリの初期化
	limiter := redis_rate.NewLimiter(rdb)
	// 信頼できるプロキシの CIDR を起動時に一度だけ解析する（tenant.go と共通）。
	trustedProxies := parseTrustedProxyCIDRs()

	return func(ctx huma.Context, next func(huma.Context)) {
		// 1. クライアントIPの特定（信頼できるプロキシ経由のヘッダーのみ採用）
		clientIP := rateLimitClientIP(ctx, trustedProxies)

		// Redisのキー: "ratelimit:<Name>:<IP>"
		// Nameが未指定の場合は "default" とする
		limitName := config.Name
		if limitName == "" {
			limitName = "default"
		}
		key := fmt.Sprintf("ratelimit:%s:%s", limitName, clientIP)

		// 2. レート制限のチェック
		// Limitオブジェクトの生成
		limit := redis_rate.Limit{
			Rate:   config.Rate,
			Period: config.Period,
			Burst:  config.Burst,
		}

		res, err := limiter.Allow(ctx.Context(), key, limit)

		// 3. Fail-Open (Redis障害時のハンドリング)
		// 可用性を優先し、Redis がダウンしていてもユーザーをブロックせず通す。
		// ただしレート制限が機能停止していることを運用者が検知できるよう必ずログを出す
		// （無言で fail-open すると、ログイン試行の制限が無効化されたことに気づけない）。
		if err != nil {
			slog.ErrorContext(ctx.Context(),
				"middleware: rate limit check failed; failing open and allowing the request",
				slog.String("limit_name", limitName),
				slog.String("error", err.Error()),
			)
			next(ctx)
			return
		}

		// 4. レート制限ヘッダーの付与 (RFC 6585 / 一般的な慣習準拠)
		// これによりクライアントは「あと何回叩けるか」を知ることができます
		ctx.SetHeader("X-RateLimit-Limit", strconv.Itoa(config.Rate))
		ctx.SetHeader("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
		ctx.SetHeader("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(res.ResetAfter).Unix(), 10))

		// 5. 制限超過時の処理
		if res.Allowed == 0 {
			// Retry-After ヘッダー (秒数)
			retryAfterSec := int(res.RetryAfter / time.Second)
			if retryAfterSec < 1 {
				retryAfterSec = 1
			}
			ctx.SetHeader("Retry-After", strconv.Itoa(retryAfterSec))

			// ヘッダーは SetStatus より前に設定する必要がある。
			// SetStatus は内部で WriteHeader を呼び出してヘッダーを確定させるため、
			// それ以降に設定したヘッダー（Content-Type 等）は破棄されてしまう。
			ctx.SetHeader("Content-Type", "application/json")

			// 429 Too Many Requests を返却
			ctx.SetStatus(http.StatusTooManyRequests)
			json.NewEncoder(ctx.BodyWriter()).Encode(map[string]any{
				"title":   "Too Many Requests",
				"status":  429,
				"detail":  "API request limit exceeded. Please try again later.",
				"message": fmt.Sprintf("Rate limit exceeded. Retry after %d seconds.", retryAfterSec),
			})
			return // next(ctx) を呼ばずに終了
		}

		// 制限内であれば次の処理へ
		next(ctx)
	}
}
