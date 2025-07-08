package config

import (
	"github.com/kelseyhightower/envconfig"
)

// New はジェネリクスを用いて、指定された型Tの構造体に環境変数から値をロードして返します。
// 戻り値:
//   - T: 環境変数からロードされた設定構造体
//   - error: ロード中に発生したエラー。正常にロードできた場合はnilを返します。
func New[T any]() (T, error) {
	v := *new(T)
	if err := envconfig.Process("", &v); err != nil {
		return *new(T), err
	}
	return v, nil
}
