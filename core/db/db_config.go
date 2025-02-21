package db

import (
	"github.com/kelseyhightower/envconfig"
)

// DBConfig はデータベース接続に必要な設定を保持する構造体です。
// 環境変数から設定値を読み込みます。
type DBConfig struct {
	// Host はデータベースサーバのホスト名です。
	Host string `envconfig:"DB_HOST" default:"localhost"`
	// Port はデータベースサーバのポート番号です。
	Port string `envconfig:"DB_PORT" default:"5432"`
	// User はデータベース接続に使用するユーザー名です。
	User string `envconfig:"DB_USER" default:"postgres"`
	// Password はデータベース接続に使用するパスワードです。
	Password string `envconfig:"DB_PASSWORD" default:"password"`
	// Database は使用するデータベース名です。
	Database string `envconfig:"DB_DATABASE" default:"sample"`
}

// Load はDBConfigの各フィールドに対して環境変数からの設定値をロードします。
func (a *DBConfig) Load() error {
	if err := envconfig.Process("", a); err != nil {
		return err
	}
	return nil
}
