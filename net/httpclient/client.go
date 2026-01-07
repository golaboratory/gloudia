package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// ClientConfig はHTTPクライアントの設定です。
type ClientConfig struct {
	Timeout      time.Duration
	MaxRetries   int
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration
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

// Client はリトライ機能とログ出力機能を備えたHTTPクライアントです。
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
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	var attempt int

	// リクエストボディのバッファリング（リトライ時に再読み込みするため）
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body.Close()
	}

	for attempt = 0; attempt <= c.config.MaxRetries; attempt++ {
		// リトライ時はボディを再設定
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		if attempt > 0 {
			wait := c.calculateBackoff(attempt)
			slog.WarnContext(req.Context(), "Retrying request",
				"attempt", attempt,
				"url", req.URL.String(),
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
				"url", req.URL.String(),
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
				"url", req.URL.String(),
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
			"url", req.URL.String(),
			"status", resp.StatusCode,
			"duration", duration,
		)
		return resp, nil
	}

	// リトライ回数超過
	return nil, fmt.Errorf("max retries reached: %w", err)
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
