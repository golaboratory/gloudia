package environment

// GloudiaEnv は Gloudia サービスの環境設定を保持する構造体です。
type GloudiaEnv struct {
	// CryptCost はパスワードハッシュ化などの暗号化処理における計算コストを指定します。
	CryptCost int  `envconfig:"CRYPT_COST"`
	IsDebug   bool `envconfig:"IS_DEBUG"`
}
