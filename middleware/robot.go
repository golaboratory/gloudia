package middleware

import (
	"net/http"
)

// NewRobotTag は全レスポンスに X-Robots-Tag: noindex, nofollow, noarchive ヘッダーを
// 付与し、検索エンジンによるインデックス登録・リンク追跡・キャッシュ保存を抑止する
// ミドルウェアを返します。
// パターン: Chi (r.Use で適用)
func NewRobotTag() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// レスポンスヘッダに noindex をセット
			w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
			next.ServeHTTP(w, r)
		})
	}
}
