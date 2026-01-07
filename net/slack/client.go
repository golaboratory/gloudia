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

// IncomingWebhookPayload はSlackへ送信するメッセージの構造体です。
type IncomingWebhookPayload struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Blocks      []Block      `json:"blocks,omitempty"`
}

// Attachment はレガシーな添付ファイルの構造体です。
type Attachment struct {
	Color  string `json:"color,omitempty"` // "good", "warning", "danger" or hex code
	Title  string `json:"title,omitempty"`
	Text   string `json:"text,omitempty"`
	Footer string `json:"footer,omitempty"`
}

// Block はBlock Kit用の簡易構造体です(詳細な定義は必要に応じて拡張)。
type Block map[string]interface{}

// Client はSlackへの通知を行うクライアントです。
type Client struct {
	webhookURL string
	httpClient *httpclient.Client
}

// NewClient は新しいSlackクライアントを作成します。
func NewClient(webhookURL string) *Client {
	return &Client{
		webhookURL: webhookURL,
		httpClient: httpclient.NewClient(httpclient.DefaultConfig()),
	}
}

// Notify は指定されたメッセージをSlackへ送信します。
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
		return fmt.Errorf("slack webhook returned status: %d", resp.StatusCode)
	}

	return nil
}

// PostText は単純なテキストメッセージを送信する簡易メソッドです。
func (c *Client) PostText(ctx context.Context, text string) error {
	return c.Notify(ctx, IncomingWebhookPayload{Text: text})
}
