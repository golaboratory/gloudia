package pdf

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/golaboratory/gloudia/net/httpclient"
	"github.com/newmo-oss/ergo"
)

// センチネルエラー定義
var (
	// ErrAPIError はGotenberg APIがエラーステータスを返した場合のエラー
	ErrAPIError = ergo.NewSentinel("api error")
)

// Client はGotenberg APIとの通信を行うクライアントです
type Client struct {
	// BaseURL は Gotenberg サーバーのベースURLです。
	BaseURL string
	// HTTPClient はリクエスト送信に使用する net/httpclient のクライアントです。
	HTTPClient *httpclient.Client
}

// NewClient は新しいGotenbergクライアントを作成します。
// HTTPClient には net/httpclient の DefaultConfig（試行ごとのタイムアウト 30 秒、
// 最大リトライ 3 回）を適用したクライアントが設定されます。
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

	// パイプを使って、multipart 作成と HTTP リクエストをストリームで繋ぐ。
	// ただし HTTPClient はリトライのためにリクエストボディを一度メモリへ
	// バッファリングする（net/httpclient 参照）。そのためピーク時のメモリ使用量は
	// ファイルサイズ相当になる点に注意（リトライ可用性を優先した設計）。
	// 非常に大きなファイルでメモリを抑えたい場合は、リトライを無効化した
	// 非バッファリングの HTTP クライアントを別途用意して送信すること。
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
			_ = pw.CloseWithError(ergo.Wrap(err, "form file creation failed"))
			return
		}
		if _, err := io.Copy(part, src); err != nil {
			_ = pw.CloseWithError(ergo.Wrap(err, "file copy failed"))
			return
		}
	}()

	// HTTPリクエストの作成
	reqURL := fmt.Sprintf("%s/forms/libreoffice/convert", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, pr)
	if err != nil {
		return nil, ergo.Wrap(err, "request creation failed")
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// リクエスト送信
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, ergo.Wrap(err, "request execution failed")
	}

	// ステータスコードのチェック
	if resp.StatusCode != http.StatusOK {
		// エラー時はbodyを読み込んでエラーメッセージとする
		resp.Body.Close()
		return nil, fmt.Errorf("%w: status %d", ErrAPIError, resp.StatusCode)
	}

	// 成功時はレスポンスボディ（PDFデータ）を返す
	// 呼び出し元で resp.Body.Close() する必要があるため、そのまま返す
	return resp.Body, nil
}
