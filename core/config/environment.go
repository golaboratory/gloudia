package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnvironments はカレントディレクトリにある ".env" ファイルから環境変数をロードします。
// 戻り値:
//   - error: 環境変数のロード中に発生したエラー。正常にロードできた場合はnilを返します。
func LoadEnvironments() error {
	return LoadEnvironmentsWithFile(".env")
}

// LoadEnvironmentsWithFile は指定されたファイルから環境変数をロードします。
// 引数:
//   - envFilePath: 読み込む環境変数ファイルのパス。ファイルが存在しない場合はエラーを返します。
//
// 戻り値:
//   - error: 環境変数のロード中に発生したエラー。正常にロードできた場合はnilを返します。
func LoadEnvironmentsWithFile(envFilePath string) error {
	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		return fmt.Errorf("環境ファイルが存在しません: %s", envFilePath)
	}

	if err := godotenv.Load(envFilePath); err != nil {
		return err
	}
	return nil
}
