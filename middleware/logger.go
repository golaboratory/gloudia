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
	env, err := environment.NewEnvValue[environment.GloudiaEnv]("")
	if err != nil {
		IsDebug = false
	}
	IsDebug = env.IsDebug
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// リクエストボディを読み取る (最大1MB)
			var bodyBytes []byte
			if IsDebug && r.Body != nil {
				// Bodyを丸ごと読むと危険なのでLimitReaderを使う
				// ただし、ログ出力のために読み切る必要があるため、上限を設けて読み込む
				const maxBodySize = 1024 * 1024 // 1MB
				reader := io.LimitReader(r.Body, maxBodySize)
				bodyBytes, _ = io.ReadAll(reader)

				// 読み取ったボディを元に戻す (読み込んだ分だけ)
				// 注: LimitReaderで止まった場合、後続の処理もそこまでしか読めない可能性があるが、
				// ここでは「ログ用」として割り切るか、あるいはBody全体をバッファする必要がある。
				// 通常のミドルウェアとしては NopCloser(bytes.NewBuffer(bodyBytes)) で戻すが、
				// IsDebug=trueは開発用前提、あるいは本番でも出すならサイズ制限必須。
				// ここでは「読み込んだデータ」だけで復元する（超過分は切り捨てられるリスクがあることに注意）
				// 本来はMultiReaderなどで繋ぐ技法があるが、シンプルに「最大1MBまでログに出す、実処理には影響させない」は難しい。
				// そのため、ここでは「最大1MBまで読み込み、それをBodyに戻す」とする。これだと1MB超のリクエストは切られることになる。
				// 安全側に倒して「1MBまで読んで、それを戻し、残りはそのまま」にするには、読み込んだbytesと元のBodyの残りを結合する必要がある。

				// 改善版: 1MBだけPeekしてログ用に使い、r.Bodyは再構築する
				// しかし io.LimitReader からは残りを読めない。
				// ここではシンプルに「1MB制限で読み込み、ログに出す。元のストリームは全て読み込まれたと仮定」する形は危険。
				// 安全策: IsDebug時のみ、1MB制限付きで読み込み、r.Bodyを「読み込んだbytes + まだ読まれていない残り」で再構築する。
				// しかし r.Body は Read してしまうと消費される。

				// アプローチ変更: バッファリングするが上限付き。
				// 上限を超えた場合はログには「truncated」とし、r.Bodyには「読み込んだ分 + 残り」として戻したいが、
				// 残りを触らずに戻すには...
				// 単純に「デバッグ時のみ」の機能と割り切り、1MB制限にする。
				r.Body = io.NopCloser(io.MultiReader(bytes.NewBuffer(bodyBytes), r.Body))
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
