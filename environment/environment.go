package environment

import (
	"log/slog"

	"github.com/kelseyhightower/envconfig"
	"github.com/newmo-oss/ergo"
)

// NewEnvValue はジェネリクスを用いて、指定された型Tの構造体に環境変数から値をロードして返します。
// 戻り値:
//   - T: 環境変数からロードされた設定構造体
//   - error: ロード中に発生したエラー。正常にロードできた場合はnilを返します。
func NewEnvValue[T any]() (T, error) {
	v := *new(T)
	if err := envconfig.Process("", &v); err != nil {
		return *new(T), ergo.New("failed to process envconfig", slog.String("error", err.Error()))
	}
	return v, nil
}
