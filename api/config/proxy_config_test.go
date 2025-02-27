package config

import (
	"os"
	"testing"

	config "github.com/golaboratory/gloudia/core/config"
	"github.com/stretchr/testify/assert"
)

// TestProxyConfig_Default は、環境変数が未設定の場合にデフォルト値が適用されることを検証します。
func TestProxyConfig_Default(t *testing.T) {
	// テスト前に環境変数をクリア
	os.Unsetenv("BACKEND_URL")
	os.Unsetenv("BINDING_PATH")

	cfg, err := config.New[ProxyConfig]()
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8000", cfg.BackendURL)
	assert.Equal(t, "/app", cfg.BindingPath)
}

// TestProxyConfig_Custom は、環境変数が設定されている場合にその値が反映されることを検証します。
func TestProxyConfig_Custom(t *testing.T) {
	os.Setenv("BACKEND_URL", "http://example.com")
	os.Setenv("BINDING_PATH", "/custom")
	defer func() {
		os.Unsetenv("BACKEND_URL")
		os.Unsetenv("BINDING_PATH")
	}()

	cfg, err := config.New[ProxyConfig]()
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com", cfg.BackendURL)
	assert.Equal(t, "/custom", cfg.BindingPath)
}
