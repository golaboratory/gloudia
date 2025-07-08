package io

import (
	"encoding/base64"
	"os"
)

// File はファイルのパスを保持する構造体です。
//   - Path: ファイルのパス
type File struct {
	Path string // ファイルのパス
}

// Base64 はBase64エンコードされたデータを保持する構造体です。
//   - Data: Base64エンコードされたデータ
type Base64 struct {
	Data string // Base64エンコードされたデータ
}

// ToBase64 はファイルの内容をBase64エンコードして文字列として返します。
// 戻り値:
//   - string: Base64エンコードされた文字列
//   - error: エンコードに失敗した場合のエラー
func (f *File) ToBase64() (string, error) {
	b, err := os.ReadFile(f.Path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// ToFile はBase64エンコードされたデータをデコードしてファイルに書き込みます。
// 戻り値:
//   - error: デコードまたは書き込みに失敗した場合のエラー
func (b *Base64) ToFile(path string) error {
	bin, err := base64.StdEncoding.DecodeString(b.Data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bin, 0644)
}
