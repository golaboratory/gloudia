package middleware

import (
	"net/http"
)

func NewRobotTag() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// レスポンスヘッダに noindex をセット
			w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
			next.ServeHTTP(w, r)
		})
	}
}
