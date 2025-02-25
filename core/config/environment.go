package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnvironments() error {
	return LoadEnvironmentsWithFile(".env")
}

func LoadEnvironmentsWithFile(path string) error {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("環境ファイルが存在しません: %s", path)
	}

	if err := godotenv.Load(path); err != nil {
		return err
	}
	return nil
}
