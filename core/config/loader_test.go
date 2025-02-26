package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// DummyConfig は、loader のテスト用構造体です。
// DUMMY_VALUE と DUMMY_COUNT の環境変数を読み込みます。
type DummyConfig struct {
	Value string `envconfig:"DUMMY_VALUE" default:"default"`
	Count int    `envconfig:"DUMMY_COUNT" default:"42"`
}

// FailConfig は、数値変換エラーを誘発するための構造体です。
// FAIL_COUNT に非数値が設定された場合にエラーとなることを検証します。
type FailConfig struct {
	Count int `envconfig:"FAIL_COUNT" default:"0"`
}

// TestNewLoader_Default は、環境変数が未設定の場合にデフォルト値が使用されることを検証します。
func TestNewLoader_Default(t *testing.T) {
	// テスト前に環境変数をクリア
	os.Unsetenv("DUMMY_VALUE")
	os.Unsetenv("DUMMY_COUNT")

	cfg, err := New[DummyConfig]()
	assert.NoError(t, err)
	assert.Equal(t, "default", cfg.Value)
	assert.Equal(t, 42, cfg.Count)
}

// TestNewLoader_WithEnv は、環境変数が設定されている場合にその値が使用されることを検証します。
func TestNewLoader_WithEnv(t *testing.T) {
	// 環境変数を設定
	os.Setenv("DUMMY_VALUE", "env_value")
	os.Setenv("DUMMY_COUNT", "100")
	defer func() {
		os.Unsetenv("DUMMY_VALUE")
		os.Unsetenv("DUMMY_COUNT")
	}()

	cfg, err := New[DummyConfig]()
	assert.NoError(t, err)
	assert.Equal(t, "env_value", cfg.Value)
	assert.Equal(t, 100, cfg.Count)
}

// TestNewLoader_Failure は、数値変換エラーが発生する場合にエラーが返されることを検証します。
func TestNewLoader_Failure(t *testing.T) {
	os.Setenv("FAIL_COUNT", "not_an_int")
	defer os.Unsetenv("FAIL_COUNT")

	_, err := New[FailConfig]()
	assert.Error(t, err)
}
