package config

import (
	"os"
	"testing"

	config "github.com/golaboratory/gloudia/core/config"
	"github.com/stretchr/testify/assert"
)

// TestStaticConfig_Default は、環境変数が設定されていない場合にデフォルト値が適用されることを検証します。
func TestStaticConfig_Default(t *testing.T) {
	// 環境変数をクリア
	os.Unsetenv("HOSTING_DIRECTORY")
	os.Unsetenv("BINDING_PATH")

	cfg, err := config.New[StaticConfig]()
	assert.NoError(t, err)
	assert.Equal(t, "./static/", cfg.HostingDirectory)
	assert.Equal(t, "/app", cfg.BindingPath)
}

// TestStaticConfig_Custom は、環境変数が設定された場合にその値が反映されることを検証します。
func TestStaticConfig_Custom(t *testing.T) {
	// カスタム値を設定
	os.Setenv("HOSTING_DIRECTORY", "/var/www")
	os.Setenv("BINDING_PATH", "/custom")
	defer func() {
		os.Unsetenv("HOSTING_DIRECTORY")
		os.Unsetenv("BINDING_PATH")
	}()

	cfg, err := config.New[StaticConfig]()
	assert.NoError(t, err)
	assert.Equal(t, "/var/www", cfg.HostingDirectory)
	assert.Equal(t, "/custom", cfg.BindingPath)
}
