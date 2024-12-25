package images

import (
	"testing"
	gqr "github.com/piglig/go-qr"
)

// TestCreateQrCodePNG はCreateQrCode関数のPNG生成をテストします。
func TestCreateQrCodePNG(t *testing.T) {
	data, err := CreateQrCode(PNG, "test text", gqr.Medium, 10, 4)
	if err != nil {
		t.Fatalf("PNG生成に失敗しました: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("生成されたPNGデータが空です")
	}
}

// TestCreateQrCodeSVG はCreateQrCode関数のSVG生成をテストします。
func TestCreateQrCodeSVG(t *testing.T) {
	data, err := CreateQrCode(SVG, "test text", gqr.Medium, 10, 4)
	if err != nil {
		t.Fatalf("SVG生成に失敗しました: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("生成されたSVGデータが空です")
	}
}

// TestCreateQrCodeUnsupportedType はCreateQrCode関数の未対応ファイルタイプをテストします。
func TestCreateQrCodeUnsupportedType(t *testing.T) {
	_, err := CreateQrCode(FileType(999), "test text", gqr.Medium, 10, 4)
	if err == nil {
		t.Fatalf("未対応ファイルタイプのエラーが発生しませんでした")
	}
}
