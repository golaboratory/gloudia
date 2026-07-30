package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/golaboratory/gloudia/environment"
)

var (
	// IsDebug は後方互換のため公開されています。NewLogger が構築時に一度だけ
	// 同期なしで書き込みます。リクエスト処理中はクロージャにキャプチャした
	// ローカル値を使用するためこの変数は読み取られませんが、ルーター構築
	// （NewLogger 呼び出し）と並行して本変数を読み書きするとデータ競合に
	// なるため避けてください。
	IsDebug = false

	// sensitiveKeyPattern は機密情報を保持するキー名の判定パターンです。
	//
	// キー名の前後に任意の語を許す部分一致で判定します。完全一致にすると
	// new_password / current_password / client_secret / password_confirmation の
	// ような派生キーが素通りしてしまうためです。過剰にマスクする方向の誤りは
	// 安全側に倒れるため許容します。
	sensitiveKeyPattern = `[a-z0-9_.\-]*(?:password|passwd|pwd|passphrase|token|secret|credential|authorization|api[_-]?key|private[_-]?key|session[_-]?id|credit[_-]?card|card[_-]?number|otp)[a-z0-9_.\-]*`

	// sensitiveJSONField は JSON ボディ中の機密フィールド値をマスクする正規表現です。
	sensitiveJSONField = regexp.MustCompile(`(?i)("` + sensitiveKeyPattern + `"\s*:\s*)"[^"]*"`)
	// sensitiveKeyValue はクエリ文字列 / フォームボディ中の機密パラメータ値をマスクする正規表現です。
	sensitiveKeyValue = regexp.MustCompile(`(?i)(` + sensitiveKeyPattern + `)=[^&\s]*`)
)

// redactSensitive は文字列中の機密情報（パスワード・トークン等）の値をマスクします。
// JSON ボディとクエリ/フォーム形式の両方に対応します。
func redactSensitive(s string) string {
	s = sensitiveJSONField.ReplaceAllString(s, `${1}"***"`)
	s = sensitiveKeyValue.ReplaceAllString(s, `${1}=***`)
	return s
}

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

// NewLogger は、アクセスログを出力するミドルウェアを返します。
// 構築時に環境変数 (GloudiaEnv) を読み取り、パッケージ変数 IsDebug を設定します。
// デバッグモードではリクエストボディも最大 1 MiB までログに出力します
// （機密情報の値はマスクされます）。
// パターン: Chi (r.Use で適用)
func NewLogger() func(http.Handler) http.Handler {
	env, err := environment.NewEnvValue[environment.GloudiaEnv]("")
	if err != nil {
		IsDebug = false
	} else {
		IsDebug = env.IsDebug
	}
	// 設定値をクロージャにキャプチャし、リクエスト処理中は共有のグローバル変数を
	// 読み取らないことでデータ競合を回避する。
	isDebug := IsDebug
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// リクエストボディを読み取る (デバッグ時のみ、最大1MB)
			var bodyBytes []byte
			if isDebug && r.Body != nil {
				const maxBodySize = 1024 * 1024 // 1MB
				reader := io.LimitReader(r.Body, maxBodySize)
				var readErr error
				bodyBytes, readErr = io.ReadAll(reader)
				if readErr != nil {
					slog.Warn("リクエストボディの読み取りに失敗", slog.String("error", readErr.Error()))
				}
				// 読み取り済みデータ + 未読の残りで Body を再構築
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

			if isDebug {
				// 機密情報（パスワード・トークン等）はマスクしてからログ出力する。
				slog.Info("Access Log",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("query", redactSensitive(r.URL.RawQuery)),
					slog.Int("status", lrw.status),
					slog.Int("size", lrw.size),
					slog.String("ip", r.RemoteAddr),
					slog.String("user_agent", r.UserAgent()),
					slog.Duration("duration", duration),
					slog.String("body", redactSensitive(string(bodyBytes))),
				)
			} else {
				slog.Info("Access Log",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("query", redactSensitive(r.URL.RawQuery)),
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
