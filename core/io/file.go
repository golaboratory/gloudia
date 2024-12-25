package io

import (
	"encoding/base64"
	"os"
)

// File は、ファイルのパスを保持する構造体です。
type File struct {
	Path string // ファイルのパス
}

// Base64 は、Base64エンコードされたデータを保持する構造体です。
type Base64 struct {
	Data string // Base64エンコードされたデータ
}

// ToBase64 は、ファイルの内容をBase64エンコードして文字列として返します。
// エンコードに失敗した場合は、エラーを返します。
func (f *File) ToBase64() (string, error) {
	b, err := os.ReadFile(f.Path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// ToFile は、Base64エンコードされたデータをデコードしてファイルに書き込みます。
// デコードまたは書き込みに失敗した場合は、エラーを返します。
func (b *Base64) ToFile(path string) error {
	bin, err := base64.StdEncoding.DecodeString(b.Data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bin, 0644)
}
