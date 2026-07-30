package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golaboratory/gloudia/net/httpclient"
	"github.com/newmo-oss/ergo"
)

// センチネルエラー定義
var (
	// ErrUnexpectedStatus はSlack Webhookが予期しないステータスコードを返した場合のエラー
	ErrUnexpectedStatus = ergo.NewSentinel("slack webhook returned unexpected status")
)

// IncomingWebhookPayload はSlackへ送信するメッセージの構造体です。
type IncomingWebhookPayload struct {
	Text        string       `json:"text"`                  // 本文テキスト (Blocks 使用時は通知等のフォールバック)
	Attachments []Attachment `json:"attachments,omitempty"` // レガシー形式の添付
	Blocks      []Block      `json:"blocks,omitempty"`      // Block Kit のブロック
}

// Attachment はレガシーな添付ファイルの構造体です。
type Attachment struct {
	Color  string `json:"color,omitempty"`  // "good", "warning", "danger" or hex code
	Title  string `json:"title,omitempty"`  // 添付のタイトル
	Text   string `json:"text,omitempty"`   // 添付の本文
	Footer string `json:"footer,omitempty"` // フッターテキスト
}

// Block はBlock Kit用の簡易構造体です(詳細な定義は必要に応じて拡張)。
type Block map[string]interface{}

// Client はSlackへの通知を行うクライアントです。
type Client struct {
	webhookURL string
	httpClient *httpclient.Client
}

// NewClient は新しいSlackクライアントを作成します。
//
// Incoming Webhook の URL はパス自体が資格情報のため、内部の HTTP クライアントには
// URL のパスをログへ出力しない設定 (RedactURLPath) を適用します。
func NewClient(webhookURL string) *Client {
	config := httpclient.DefaultConfig()
	config.RedactURLPath = true
	return &Client{
		webhookURL: webhookURL,
		httpClient: httpclient.NewClient(config),
	}
}

// Notify は指定されたメッセージをSlackへ送信します。
// 内部で net/httpclient (DefaultConfig: 試行ごと 30 秒タイムアウト、最大 3 回リトライ)
// を使用します。応答が 200 以外の場合は ErrUnexpectedStatus をラップしたエラーを返します。
func (c *Client) Notify(ctx context.Context, payload IncomingWebhookPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return ergo.Wrap(err, "failed to marshal payload")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return ergo.Wrap(err, "failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")

	// 共通の堅牢なHTTPクライアントを使用（リトライ等も自動）
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ergo.Wrap(err, "failed to posting to slack")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %d", ErrUnexpectedStatus, resp.StatusCode)
	}

	return nil
}

// PostText は単純なテキストメッセージを送信する簡易メソッドです。
func (c *Client) PostText(ctx context.Context, text string) error {
	return c.Notify(ctx, IncomingWebhookPayload{Text: text})
}
