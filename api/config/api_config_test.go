package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDefaultApiConfig は環境変数が設定されていない場合、ApiConfig.Loadがデフォルト値をセットすることを検証します。
func TestDefaultApiConfig(t *testing.T) {
	// テスト前に環境変数をクリア
	os.Clearenv()

	var cfg ApiConfig
	err := cfg.Load()
	assert.NoError(t, err)

	// デフォルト値の検証
	assert.Equal(t, 8888, cfg.Port)
	assert.Equal(t, true, cfg.EnableJWT)
	assert.Equal(t, true, cfg.EnableStatic)
	assert.Equal(t, false, cfg.EnableSSL)
	assert.Equal(t, true, cfg.EnableCookieToken)
	assert.Equal(t, "/api", cfg.RootPath)
	assert.Equal(t, "Sample API", cfg.APITitle)
	assert.Equal(t, "1.0.0", cfg.APIVersion)
	assert.Equal(t, "BHqQTg99LmSk$Q,_xe*LM+!P*5PKnR~n", cfg.JWTSecret)
	assert.Equal(t, 480, cfg.JWTExpireMinute)
}

// TestCustomApiConfig は環境変数を設定した場合、ApiConfig.Loadが各フィールドに正しい値を反映することを検証します。
func TestCustomApiConfig(t *testing.T) {
	// 環境変数を設定
	os.Setenv("PORT", "9999")
	os.Setenv("ENABLE_JWT", "false")
	os.Setenv("ENABLE_STATIC", "false")
	os.Setenv("ENABLE_SSL", "true")
	os.Setenv("ENABLE_COOKIE_TOKEN", "false")
	os.Setenv("ROOT_PATH", "/custom")
	os.Setenv("API_TITLE", "Custom API")
	os.Setenv("API_VERSION", "2.0.0")
	os.Setenv("JWT_SECRET", "CustomSecret")
	os.Setenv("JWT_EXPIRE", "60")

	// テスト終了時に環境変数をリセット
	t.Cleanup(func() {
		os.Clearenv()
	})

	var cfg ApiConfig
	err := cfg.Load()
	assert.NoError(t, err)

	// 環境変数で設定した値の検証
	assert.Equal(t, 9999, cfg.Port)
	assert.Equal(t, false, cfg.EnableJWT)
	assert.Equal(t, false, cfg.EnableStatic)
	assert.Equal(t, true, cfg.EnableSSL)
	assert.Equal(t, false, cfg.EnableCookieToken)
	assert.Equal(t, "/custom", cfg.RootPath)
	assert.Equal(t, "Custom API", cfg.APITitle)
	assert.Equal(t, "2.0.0", cfg.APIVersion)
	assert.Equal(t, "CustomSecret", cfg.JWTSecret)
	assert.Equal(t, 60, cfg.JWTExpireMinute)
}
