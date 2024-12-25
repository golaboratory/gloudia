package images

import (
	"fmt"
	"os"

	gqr "github.com/piglig/go-qr"
)

// FileType はファイルの種類を表します。
type FileType int

const (
	// PNG はPNGファイルを表します。
	PNG FileType = iota
	// SVG はSVGファイルを表します。
	SVG
)

// CreateQrCode はQRコードを生成し、指定されたファイル形式で返却します。
// ft: ファイルの種類（PNGまたはSVG）
// text: QRコードにエンコードするテキスト
// level: エラー訂正レベル
// scale: スケール
// border: ボーダーサイズ
// 戻り値: 生成されたQRコードのバイトデータとエラー情報
func CreateQrCode(ft FileType, text string, level gqr.Ecc, scale int, border int) (data []byte, err error) {
	qr, err := gqr.EncodeText(text, level)
	if err != nil {
		return []byte(""), err
	}
	config := gqr.NewQrCodeImgConfig(scale, border)

	f, err := os.CreateTemp("", "gloudia.core.images.qr")
	if err != nil {
		return []byte(""), err
	}
	defer os.Remove(f.Name())

	switch ft {
	case PNG:
		err = qr.PNG(config, f.Name())
	case SVG:
		err = qr.SVG(config, f.Name(), "#FFFFFF", "#000000")
	default:
		return []byte(""), fmt.Errorf("unsupported file type")
	}
	if err != nil {
		return []byte(""), err
	}

	data, err = os.ReadFile(f.Name())
	if err != nil {
		return []byte(""), err
	}

	return data, nil
}
