package environment

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/newmo-oss/ergo"
)

// センチネルエラー定義
var (
	// ErrEnvConfigProcessing は環境変数の読み込み処理に失敗した場合のエラー
	ErrEnvConfigProcessing = ergo.NewSentinel("failed to process envconfig")
)

// NewEnvValue はジェネリクスを用いて、指定された型Tの構造体に環境変数から値をロードして返します。
// prefix を指定すると、そのプレフィックスを持つ環境変数のみ読み込みます（例: "APP" -> APP_VAR_NAME）。
// 空文字の場合はトップレベルの環境変数を探します。
// 戻り値:
//   - T: 環境変数からロードされた設定構造体
//   - error: ロード中に発生したエラー。正常にロードできた場合はnilを返します。
func NewEnvValue[T any](prefix string) (T, error) {
	v := *new(T)
	if err := envconfig.Process(prefix, &v); err != nil {
		return *new(T), ergo.Wrap(ErrEnvConfigProcessing, err.Error())
	}
	return v, nil
}
