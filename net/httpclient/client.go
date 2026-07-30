package httpclient

import (
	"bytes"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/newmo-oss/ergo"
)

// センチネルエラー定義
var (
	// ErrMaxRetriesExceeded は、HTTP 5xx 応答のままリトライ回数の上限に達した
	// 場合に返されるエラー。最後の試行がネットワーク等のトランスポートエラーで
	// 失敗した場合は、そのエラーをラップした別のエラー（メッセージ
	// "max retries reached"）が返り、errors.Is(err, ErrMaxRetriesExceeded) には
	// 一致しないことに注意。
	ErrMaxRetriesExceeded = ergo.NewSentinel("max retries reached")
)

// sensitiveQueryKey はログ出力時に値をマスクするクエリパラメータ名のパターンです。
// キー名の前後に任意の語を許す部分一致で判定します。
var sensitiveQueryKey = regexp.MustCompile(`(?i)^[a-z0-9_.\-]*(?:password|passwd|pwd|passphrase|token|secret|credential|signature|api[_-]?key|access[_-]?key|private[_-]?key|session[_-]?id|otp|sig)[a-z0-9_.\-]*$`)

// redactURL はログ出力用に URL から機密情報を除去した文字列を返します。
//   - ユーザー情報 (user:password@host) を除去します
//   - クエリパラメータのうち機密キーに一致する値を *** に置換します
//   - フラグメントを除去します
//   - redactPath が true の場合はパスとクエリ全体を /redacted に置換します
//
// URL のパス自体が資格情報となるエンドポイント（Slack Incoming Webhook など）では
// redactPath を有効にしてください。
func redactURL(u *url.URL, redactPath bool) string {
	if u == nil {
		return ""
	}
	c := *u
	c.User = nil
	c.Fragment = ""

	if redactPath {
		// url.URL.String() は "*" をパーセントエンコードするため、
		// ログで読みにくくならないマーカーを使う。
		c.Path = "/redacted"
		c.RawPath = ""
		c.RawQuery = ""
		return c.String()
	}

	if c.RawQuery != "" {
		q := c.Query()
		for k, vs := range q {
			if sensitiveQueryKey.MatchString(k) {
				for i := range vs {
					vs[i] = "***"
				}
			}
		}
		c.RawQuery = q.Encode()
	}
	return c.String()
}

// ClientConfig はHTTPクライアントの設定です。
type ClientConfig struct {
	// Timeout は 1 回の試行ごとのタイムアウトです。リトライを含めた最悪の所要時間は
	// おおよそ Timeout×(MaxRetries+1) + バックオフ待機の合計になります。
	Timeout time.Duration
	// MaxRetries はリトライ回数です。MaxRetries=3 の場合、最大 4 回試行します。
	MaxRetries int
	// RetryWaitMin は指数バックオフの初期待機時間です。
	RetryWaitMin time.Duration
	// RetryWaitMax はバックオフ待機時間の上限です。上限適用後に ±10% のジッタが
	// 掛かるため、実際の待機時間は最大で RetryWaitMax×1.1 になり得ます。
	RetryWaitMax time.Duration

	// RedactURLPath を true にすると、ログ出力時に URL のパスもマスクします。
	// Slack Incoming Webhook のように URL のパス自体が資格情報となる
	// エンドポイントへ送信する場合に有効にしてください。
	// 既定 (false) では、ユーザー情報と機密クエリパラメータのみをマスクします。
	RedactURLPath bool
}

// DefaultConfig はデフォルトの設定を返します。
func DefaultConfig() ClientConfig {
	return ClientConfig{
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		RetryWaitMin: 1 * time.Second,
		RetryWaitMax: 5 * time.Second,
	}
}

// HTTPDoer はHTTPリクエストを実行するインターフェースです。
// *http.Client および本パッケージの *Client がこのインターフェースを満たします
// （Do はポインタレシーバのため、値型の Client は満たしません）。
// テスト時にモックへ差し替えることで、外部通信なしにHTTPクライアント利用コードを検証できます。
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client はリトライ機能とログ出力機能を備えたHTTPクライアントです。
// *Client が HTTPDoer インターフェースを満たします。
type Client struct {
	client *http.Client
	config ClientConfig
}

// NewClient は新しいHTTPクライアントを作成します。
func NewClient(config ClientConfig) *Client {
	return &Client{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
	}
}

// Do はHTTPリクエストを実行します。
// 500系エラーやネットワークエラーの場合、設定に基づいてリトライを行います。
// リトライのためリクエストボディは全量メモリにバッファリングされ、req.Body は
// 差し替えられます（ストリーミングボディは非対応）。
// 4xx 応答はリトライせず err=nil でそのまま返します。5xx のままリトライ上限に
// 達した場合は最後のレスポンスを破棄し ErrMaxRetriesExceeded を返します。
// バックオフ待機中に context がキャンセルされた場合は ctx.Err() を返します。
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	var attempt int

	// リクエストボディのバッファリング（リトライ時に再読み込みするため）
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, ergo.Wrap(err, "failed to read request body")
		}
		req.Body.Close()
	}

	// ログ出力用にマスク済みの URL を一度だけ組み立てる。
	// 生の URL は資格情報（ユーザー情報・クエリのトークン・Webhook のパス）を
	// 含みうるため、ログにはこのマスク済み文字列のみを出力する。
	logURL := redactURL(req.URL, c.config.RedactURLPath)

	for attempt = 0; attempt <= c.config.MaxRetries; attempt++ {
		// リトライ時はボディを再設定
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		if attempt > 0 {
			wait := c.calculateBackoff(attempt)
			slog.WarnContext(req.Context(), "Retrying request",
				"attempt", attempt,
				"url", logURL,
				"wait", wait,
			)
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(wait):
			}
		}

		start := time.Now()
		resp, err = c.client.Do(req)
		duration := time.Since(start)

		if err != nil {
			slog.ErrorContext(req.Context(), "Request failed",
				"method", req.Method,
				"url", logURL,
				"error", err,
				"duration", duration,
			)
			// ネットワークエラーなどはリトライ対象
			continue
		}

		// ステータスコードチェック
		if resp.StatusCode >= 500 {
			slog.WarnContext(req.Context(), "Server error",
				"method", req.Method,
				"url", logURL,
				"status", resp.StatusCode,
				"duration", duration,
			)
			resp.Body.Close() // 次のリトライの前に閉じる
			// 500系はリトライ対象
			continue
		}

		// 成功 (2xx - 4xx)
		slog.DebugContext(req.Context(), "Request finished",
			"method", req.Method,
			"url", logURL,
			"status", resp.StatusCode,
			"duration", duration,
		)
		return resp, nil
	}

	// リトライ回数超過
	if err != nil {
		return nil, ergo.Wrap(err, "max retries reached")
	}
	return nil, ErrMaxRetriesExceeded
}

// calculateBackoff は指数バックオフ時間を計算します (Jitter付き)。
func (c *Client) calculateBackoff(attempt int) time.Duration {
	// 2^attempt * min
	base := float64(c.config.RetryWaitMin) * math.Pow(2, float64(attempt-1))

	// Cap at max
	if base > float64(c.config.RetryWaitMax) {
		base = float64(c.config.RetryWaitMax)
	}

	// Jitter: +/- 10%
	jitter := (rand.Float64() * 0.2) + 0.9 // 0.9 ~ 1.1
	return time.Duration(base * jitter)
}

// Helper Wrappers (Get, Post, etc.) could be added here similar to http.Client
