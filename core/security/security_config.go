package security

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// SecurityConfig はセキュリティに必要なパスワードおよびソルトなどの設定を管理する型です。
//   - Password: 暗号化やハッシュ化に使用するパスワード
//   - Salt: 暗号化やハッシュ化に使用するソルト
type SecurityConfig struct {
	Password string `envconfig:"SECURITY_PASSWORD" default:"password"`
	Salt     string `envconfig:"SECURITY_SALT" default:"salt"`
}

// Load は環境変数からセキュリティ設定を読み込み、SecurityConfigを初期化します。
// エラーが発生した場合はそのエラーを返します。
func (a *SecurityConfig) Load() error {
	if err := envconfig.Process("", a); err != nil {
		return err
	}
	return nil
}

// NewSecurityConfig はSecurityConfigの新しいインスタンスを生成し、環境変数から設定をロードします。
// ロード時にエラーが発生した場合はエラー内容を標準出力に表示し、デフォルト値で初期化された設定を返します。
func NewSecurityConfig() *SecurityConfig {
	sec := &SecurityConfig{}
	err := sec.Load()
	if err != nil {
		fmt.Println(err)
	}
	return sec
}
