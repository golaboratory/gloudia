package config

// ProxyConfig はリバースプロキシの設定情報を保持する構造体です。
// 環境変数から設定値を読み込み、バックエンドURLおよびバインディングパスを指定します。
type ProxyConfig struct {
	// BackendURL はリバースプロキシが転送するバックエンドサーバのURLを指定します。
	BackendURL string `envconfig:"BACKEND_URL" default:"http://localhost:8000"`

	// BindingPath はリバースプロキシがバインドされるパスを指定します。
	BindingPath string `envconfig:"BINDING_PATH" default:"/app"`
}
