package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client はHTTPクライアントを表します。
type Client struct {
	UseBearerToken bool         // ベアラートークンを使用するかどうか
	BearerToken    string       // ベアラートークン
	Timeout        int          // タイムアウト時間（秒）
	RetryCount     int          // リトライ回数
	HttpClient     *http.Client // HTTPクライアント
}

// Init はHTTPクライアントを初期化します。
func (c *Client) init() {
	var retry = 1
	if c.RetryCount > 1 {
		retry = c.RetryCount
	}

	c.HttpClient = &http.Client{
		Timeout: time.Duration(c.Timeout) * time.Second,
		Transport: &GloudiaRoundTripper{
			t:        http.DefaultTransport,
			maxRetry: retry,
			wait:     2 * time.Second,
		},
	}
}

// GetString は指定されたURLから文字列を取得します。
func (c *Client) GetString(url string) (string, error) {
	arr, err := c.Get(url)
	if err != nil {
		return "", err
	}
	return string(arr), nil
}

// GetFile は指定されたURLからファイルを取得し、一時ファイルとして保存します。
func (c *Client) GetFile(url string) (string, error) {
	arr, err := c.Get(url)
	if err != nil {
		return "", err
	}
	f, err := os.CreateTemp("", "gloudia.core.http.get")
	if err != nil {
		return "", err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			fmt.Printf("failed to remove file: %v\n", err)
		}
	}()

	_, err = f.Write(arr)
	if err != nil {
		return "", err
	}

	return f.Name(), nil
}

// Get は指定されたURLからデータを取得します。
func (c *Client) Get(url string) ([]byte, error) {

	c.init()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return []byte(""), err
	}

	if c.UseBearerToken {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.BearerToken))
	}

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return []byte(""), err
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}()

	byteArray, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte(""), err
	}

	return byteArray, nil
}

// PostJson は指定されたURLにJSONデータをPOSTし、レスポンスを文字列として返します。
func (c *Client) PostJson(url string, json string) (string, error) {
	arr, err := c.Post(url, []byte(json), "application/json")
	if err != nil {
		return "", err
	}
	return string(arr), nil
}

// Post は指定されたURLにデータをPOSTし、レスポンスをバイト配列として返します。
func (c *Client) Post(url string, data []byte, contentType string) ([]byte, error) {
	c.init()

	var ct = "application/x-www-form-urlencoded"
	if contentType != "" {
		ct = contentType
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return []byte(""), err
	}

	if c.UseBearerToken {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.BearerToken))
	}

	req.Header.Set("Content-Type", ct)

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return []byte(""), err
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}()

	byteArray, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte(""), err
	}

	return byteArray, nil
}
