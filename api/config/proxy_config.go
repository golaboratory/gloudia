package config

// ProxyConfig はリバースプロキシの設定情報を保持する構造体です。
// 各フィールドは環境変数から設定値を読み込みます。
type ProxyConfig struct {
	// BackendURL はリバースプロキシが転送するバックエンドサーバのURLです。
	BackendURL string `envconfig:"BACKEND_URL" default:"http://localhost:8000"`
	// BindingPath はリバースプロキシがバインドされるパスです。
	BindingPath string `envconfig:"BINDING_PATH" default:"/app"`
}
