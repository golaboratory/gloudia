package middleware

import (
	"github.com/go-chi/cors"
)

// NewCORS は Chi ルーター用の CORS 設定を返します。
// main.go で router.Use(middleware.NewCORS()) のように使用します。
func NewCORS() *cors.Cors {
	return cors.New(cors.Options{
		// 開発環境と本番環境で許可オリジンを切り替える必要があります。
		// 環境変数から取得することを推奨します。
		AllowedOrigins: []string{"https://*", "http://localhost:5173"}, // Vite Default Port

		AllowedMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Forwarded-Host", // テナント指定用
		},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Preflightリクエストのキャッシュ時間
	})
}
