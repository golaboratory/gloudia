package config

// StaticConfig は静的ファイルのホスティング設定を保持する構造体です。
// 環境変数から値をロードし、ホスティングディレクトリとバインディングパスを指定します。
type StaticConfig struct {
	// HostingDirectory は静的ファイルが配置されるディレクトリのパスです。
	HostingDirectory string `envconfig:"HOSTING_DIRECTORY" default:"./static/"`
	// BindingPath は静的ファイルの提供に使用されるエンドポイントのパスです。
	BindingPath string `envconfig:"BINDING_PATH" default:"/app"`
}
