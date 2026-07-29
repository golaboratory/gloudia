package middleware

import (
	"os"
	"strings"

	"github.com/go-chi/cors"
)

// NewCORS は Chi ルーター用の CORS 設定を返します。
// main.go で router.Use(middleware.NewCORS()) のように使用します。
//
// セキュリティ上の理由から、任意の HTTPS オリジンを反射する "https://*" の
// ワイルドカードは使用しない（AllowCredentials:true と併用すると、任意の悪意ある
// サイトが資格情報付きクロスオリジンリクエストを実行できてしまうため）。
//
// 既定の許可オリジンは "http://localhost:5173"（Vite 既定ポート）のみ。
// 本番ドメインを含むその他の許可オリジンは、必ず環境変数 CORS_ALLOWED_ORIGINS
// （カンマ区切り）で明示的に指定すること。
// 例: CORS_ALLOWED_ORIGINS="https://app.example.com,https://admin.example.com"
// ホストサフィックスのワイルドカード（例: "https://*.example.com"）も指定可能。
func NewCORS() *cors.Cors {
	allowedOrigins := []string{"http://localhost:5173"} // Vite Default Port (開発用)
	if extra := os.Getenv("CORS_ALLOWED_ORIGINS"); strings.TrimSpace(extra) != "" {
		for _, origin := range strings.Split(extra, ",") {
			if o := strings.TrimSpace(origin); o != "" {
				allowedOrigins = append(allowedOrigins, o)
			}
		}
	}

	return cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,

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
