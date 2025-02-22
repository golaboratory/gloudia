package security

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// SecurityConfig は、セキュリティに必要なパスワードおよびソルトなどの設定を管理する型です。
type SecurityConfig struct {
	Password string `envconfig:"SECURITY_PASSWORD" default:"password"`
	Salt     string `envconfig:"SECURITY_SALT" default:"salt"`
}

// Load は、環境変数からセキュリティ設定を読み込み、SecurityConfig を初期化します。
// エラーが発生した場合は、そのエラーを返します。
func (a *SecurityConfig) Load() error {
	if err := envconfig.Process("", a); err != nil {
		return err
	}
	return nil
}

// NewSecurityConfig は、SecurityConfig の新しいインスタンスを生成し、環境変数から設定をロードします。
// ロード時にエラーが発生した場合は、エラー内容を標準出力に表示し、デフォルト値で初期化された設定を返します。
func NewSecurityConfig() *SecurityConfig {
	sec := &SecurityConfig{}
	err := sec.Load()
	if err != nil {
		fmt.Println(err)
	}
	return sec
}
