package db

// ValkeyConfig はValkeyサーバ接続に必要な設定を保持する構造体です。
// 各フィールドは環境変数から設定値を読み込みます。
type ValkeyConfig struct {
	Host string `envconfig:"DB_HOST" default:"localhost"`
	Port string `envconfig:"DB_PORT" default:"15432"`
}
