package config

import (
	"github.com/kelseyhightower/envconfig"
)

// ApiConfig はAPIサーバの設定項目を保持する構造体です。
// 環境変数から設定値を読み込みます。
type ApiConfig struct {
	// Port はサーバがリッスンするポート番号です。
	Port int `envconfig:"PORT" default:"8888"`
	// EnableJWT はJWT認証を有効にするか否かを示します。
	EnableJWT bool `envconfig:"ENABLE_JWT" default:"true"`
	// EnableStatic は静的ファイルの提供を有効にするか否かを示します。
	EnableStatic bool `envconfig:"ENABLE_STATIC" default:"true"`
	// EnableSSL はSSL通信を有効にするか否かを示します。
	EnableSSL bool `envconfig:"ENABLE_SSL" default:"false"`
	// EnableCookieToken はCookieからのJWTトークン取得を有効にするかを示します。
	EnableCookieToken bool `envconfig:"ENABLE_COOKIE_TOKEN" default:"true"`
	// RootPath はAPIのルートパスです。
	RootPath string `envconfig:"ROOT_PATH" default:"/api"`
	// APITitle はAPIのタイトルです。
	APITitle string `envconfig:"API_TITLE" default:"Sample API"`
	// APIVersion はAPIのバージョンを示します。
	APIVersion string `envconfig:"API_VERSION" default:"1.0.0"`
	// JWTSecret はJWT署名検証に使用されるシークレットキーです。
	JWTSecret string `envconfig:"JWT_SECRET" default:"BHqQTg99LmSk$Q,_xe*LM+!P*5PKnR~n"`
	// JWTExpireMinute はJWTトークンの有効期限（分）です。
	JWTExpireMinute int `envconfig:"JWT_EXPIRE" default:"480"`
}

// Load はApiConfigの各フィールドに対して環境変数からの設定値をロードします。
func (a *ApiConfig) Load() error {
	if err := envconfig.Process("", a); err != nil {
		return err
	}
	return nil
}
