package environment

import (
	"log/slog"

	"github.com/kelseyhightower/envconfig"
	"github.com/newmo-oss/ergo"
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
		return *new(T), ergo.New("failed to process envconfig", slog.String("error", err.Error()))
	}
	return v, nil
}
