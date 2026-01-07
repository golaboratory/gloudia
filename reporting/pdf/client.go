package pdf

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/golaboratory/gloudia/net/httpclient"
)

// Client はGotenberg APIとの通信を行うクライアントです
type Client struct {
	BaseURL    string
	HTTPClient *httpclient.Client
}

// NewClient は新しいGotenbergクライアントを作成します
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: httpclient.NewClient(httpclient.DefaultConfig()),
	}
}

// Convert はExcel(io.Reader)を受け取り、PDF(io.ReadCloser)を返します。
// 呼び出し元は、返却されたReadCloserを必ずCloseする必要があります。
func (c *Client) Convert(ctx context.Context, filename string, src io.Reader, opts *ConvertOptions) (io.ReadCloser, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// パイプを使って、multipart作成とHTTPリクエストをストリームで繋ぐ
	// これにより、大きなファイルでもメモリを圧迫しにくくなります
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// ゴルーチンでmultipartデータを書き込む
	go func() {
		defer pw.Close()
		defer writer.Close()

		// 1. オプションパラメータの書き込み
		if opts.Landscape {
			_ = writer.WriteField("landscape", "true")
		}
		if opts.PageRanges != "" {
			_ = writer.WriteField("nativePageRanges", opts.PageRanges)
		}
		if opts.Scale > 0 {
			_ = writer.WriteField("scale", fmt.Sprintf("%.2f", opts.Scale))
		}

		// 2. ファイルデータの書き込み
		// form-dataのキーは "files" である必要があります
		part, err := writer.CreateFormFile("files", filename)
		if err != nil {
			_ = pw.CloseWithError(fmt.Errorf("form file creation failed: %w", err))
			return
		}
		if _, err := io.Copy(part, src); err != nil {
			_ = pw.CloseWithError(fmt.Errorf("file copy failed: %w", err))
			return
		}
	}()

	// HTTPリクエストの作成
	reqURL := fmt.Sprintf("%s/forms/libreoffice/convert", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, pr)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// リクエスト送信
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request execution failed: %w", err)
	}

	// ステータスコードのチェック
	if resp.StatusCode != http.StatusOK {
		// エラー時はbodyを読み込んでエラーメッセージとする
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(body))
	}

	// 成功時はレスポンスボディ（PDFデータ）を返す
	// 呼び出し元で resp.Body.Close() する必要があるため、そのまま返す
	return resp.Body, nil
}
