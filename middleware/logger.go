package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/golaboratory/gloudia/environment"
)

var (
	IsDebug = false
)

// accessLogResponseWriter は、ステータスコードとレスポンスサイズをキャプチャするためのラッパーです
type accessLogResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *accessLogResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *accessLogResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

// NewLogger は、アクセスログを出力するミドルウェアを返します
func NewLogger() func(http.Handler) http.Handler {
	env, err := environment.NewEnvValue[environment.GloudiaEnv]()
	if err != nil {
		IsDebug = false
	}
	IsDebug = env.IsDebug
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// リクエストボディを読み取る
			var bodyBytes []byte
			if IsDebug && r.Body != nil {
				bodyBytes, _ = io.ReadAll(r.Body)
				// 読み取ったボディを元に戻す
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			// ステータスコードキャプチャ用のラッパーを作成
			lrw := &accessLogResponseWriter{ResponseWriter: w}

			next.ServeHTTP(lrw, r)

			// ステータスが設定されていない場合は200とみなす
			if lrw.status == 0 {
				lrw.status = http.StatusOK
			}

			duration := time.Since(start)

			if IsDebug {
				slog.Info("Access Log",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("query", r.URL.RawQuery),
					slog.Int("status", lrw.status),
					slog.Int("size", lrw.size),
					slog.String("ip", r.RemoteAddr),
					slog.String("user_agent", r.UserAgent()),
					slog.Duration("duration", duration),
					slog.String("body", string(bodyBytes)),
				)
			} else {
				slog.Info("Access Log",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("query", r.URL.RawQuery),
					slog.Int("status", lrw.status),
					slog.Int("size", lrw.size),
					slog.String("ip", r.RemoteAddr),
					slog.String("user_agent", r.UserAgent()),
					slog.Duration("duration", duration),
				)
			}
		})
	}
}
