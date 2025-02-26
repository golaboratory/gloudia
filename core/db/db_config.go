package db

// DBConfig はデータベース接続に必要な設定を保持する構造体です。
// 環境変数から設定値を読み込みます。
type DBConfig struct {
	// Host はデータベースサーバのホスト名です。
	Host string `envconfig:"DB_HOST" default:"localhost"`
	// Port はデータベースサーバのポート番号です。
	Port string `envconfig:"DB_PORT" default:"15432"`
	// User はデータベース接続に使用するユーザー名です。
	User string `envconfig:"DB_USER" default:"postgres"`
	// Password はデータベース接続に使用するパスワードです。
	Password string `envconfig:"DB_PASSWORD" default:"postgres"`
	// Database は使用するデータベース名です。
	Database string `envconfig:"DB_DATABASE" default:"postgres"`
}
