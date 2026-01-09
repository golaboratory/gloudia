package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig レート制限の設定構造体
type RateLimitConfig struct {
	Rate   int           // 期間あたりの許可リクエスト数
	Burst  int           // バースト（瞬間的な許容超過数）
	Period time.Duration // 期間 (例: 1秒, 1分)
	Name   string        // レートリミット識別子 (例: "global", "login")
}

// NewRedisRateLimiter ミドルウェアを生成するファクトリ関数
func NewRedisRateLimiter(rdb *redis.Client, config RateLimitConfig) func(huma.Context, func(huma.Context)) {
	// redis_rate ライブラリの初期化
	limiter := redis_rate.NewLimiter(rdb)

	limitNameInit := config.Name
	if limitNameInit == "" {
		limitNameInit = "default"
	}
	// 初期化時に既存のレートリミット情報をクリアする処理は削除 (他インスタンスへの影響回避のため)

	return func(ctx huma.Context, next func(huma.Context)) {
		// 1. クライアントIPの特定 (Nginx考慮)
		// 統合仕様書のNginx設定にある "proxy_set_header X-Real-IP $remote_addr;" を信頼します
		clientIP := ctx.Header("X-Real-IP")
		if clientIP == "" {
			// ヘッダーがない場合はX-Forwarded-Forを確認
			xff := ctx.Header("X-Forwarded-For")
			if xff != "" {
				parts := strings.Split(xff, ",")
				clientIP = strings.TrimSpace(parts[0])
			}
		}
		if clientIP == "" {
			// それでも取れなければ直接接続元のIP (開発環境など)
			clientIP = ctx.RemoteAddr()
		}

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
		if err != nil {
			// Redisがダウンしていても、ユーザーをブロックせず通す (ログ出力推奨)
			// logger.Error("Redis rate limit error", "error", err)
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

			// 429 Too Many Requests を返却
			ctx.SetStatus(http.StatusTooManyRequests)
			ctx.SetHeader("Content-Type", "application/json")
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
