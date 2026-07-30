package environment

// GloudiaEnv は Gloudia サービスの環境設定を保持する構造体です。
type GloudiaEnv struct {
	// CryptCost はパスワードハッシュ化などの暗号化処理における計算コストを指定します。
	// auth.HashPassword で参照され、bcrypt の許容範囲（bcrypt.MinCost〜bcrypt.MaxCost、4〜31）外の
	// 値は無視されて bcrypt.DefaultCost が使用されます。
	CryptCost int `envconfig:"CRYPT_COST"`
	// IsDebug は環境変数 IS_DEBUG から読み込まれ、
	// middleware のデバッグ用リクエストログ出力を有効にします。
	IsDebug bool `envconfig:"IS_DEBUG"`
}
