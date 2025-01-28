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

// Create はQRコードを生成し、指定されたファイル形式で返却します。
// ft: ファイルの種類（PNGまたはSVG）
// text: QRコードにエンコードするテキスト
// level: エラー訂正レベル
// scale: スケール
// border: ボーダーサイズ
// 戻り値: 生成されたQRコードのバイトデータとエラー情報
func Create(ft FileType, text string, level gqr.Ecc, scale int, border int) (data []byte, err error) {

	if ft != PNG && ft != SVG {
		return []byte(""), fmt.Errorf("unsupported file type")
	}

	qr, err := gqr.EncodeText(text, level)
	if err != nil {
		return []byte(""), err
	}
	config := gqr.NewQrCodeImgConfig(scale, border)

	f, err := os.CreateTemp("", "gloudia.core.images.qr")
	if err != nil {
		return []byte(""), err
	}

	var orgDestPath = f.Name()
	err = f.Close()
	if err != nil {
		return []byte(""), err
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Printf("failed to remove file: %v\n", err)
		}
	}(orgDestPath)

	var ext = ".png"
	if ft == SVG {
		ext = ".svg"
	}

	var destPath = orgDestPath + ext
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Printf("failed to remove file: %v\n", err)
		}
	}(destPath)

	if ft == PNG {
		err = qr.PNG(config, destPath)
	} else {
		err = qr.SVG(config, destPath, "#FFFFFF", "#000000")
	}

	if err != nil {
		return []byte(""), err
	}

	data, err = os.ReadFile(destPath)
	if err != nil {
		return []byte(""), err
	}

	return data, nil
}
