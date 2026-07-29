package pdf

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/newmo-oss/ergo"
)

// Converter はファイル変換処理を行うための構造体です。
type Converter struct {
	excelPath    string
	pdfPath      string
	gotenbergUrl string
	option       *ConvertOptions
	timeout      time.Duration
}

// NewConverter は新しい Converter インスタンスを作成します。
// excelPath: 変換元の Excel ファイルパス
// pdfPath: 変換後の PDF 出力先パス
// gotenbergUrl: Gotenberg サーバーの URL
// option: 変換オプション (nil の場合はデフォルト設定が使用される可能性があります)
func NewConverter(excelPath string, pdfPath string, gotenbergUrl string, option *ConvertOptions) *Converter {
	return &Converter{
		excelPath:    excelPath,
		pdfPath:      pdfPath,
		gotenbergUrl: gotenbergUrl,
		option:       option,
		timeout:      30 * time.Second,
	}
}

// FromExcel は Excel ファイルを PDF に変換します。
// 変換された PDF を pdfPath に保存し、そのパスを返します。
// 処理には 30秒のタイムアウトが設定されています。
func (c *Converter) FromExcel() (string, error) {
	// 1. クライアントの初期化
	// Dockerで起動しているGotenbergのURLを指定
	client := NewClient(c.gotenbergUrl)

	// 2. 入力ファイルを開く (io.Readerであれば何でもOK)
	excelFile, err := os.Open(c.excelPath)
	if err != nil {
		return "", ergo.Wrap(err, "failed to open input file")
	}
	defer excelFile.Close()

	// 3. オプション設定 (例えば横向き、1ページ目のみ)
	opts := c.option

	// 4. 変換実行 (タイムアウト設定付き)
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Convertメソッドは io.ReadCloser を返す
	pdfStream, err := client.Convert(ctx, c.excelPath, excelFile, opts)
	if err != nil {
		return "", ergo.Wrap(err, "failed to convert")
	}
	// 必ずCloseする
	defer pdfStream.Close()

	// 5. 結果をファイルに保存
	outFile, err := os.Create(c.pdfPath)
	if err != nil {
		return "", ergo.Wrap(err, "failed to create output file")
	}
	defer outFile.Close()

	// ストリームとしてコピー（メモリ効率が良い）
	_, err = io.Copy(outFile, pdfStream)
	if err != nil {
		return "", ergo.Wrap(err, "failed to copy stream")
	}
	return c.pdfPath, nil
}
