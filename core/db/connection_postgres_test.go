package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewPostgresConnection_Failure は、DBConfigの内容が正しくない場合に接続が失敗することを検証します。
func TestNewPostgresConnection_Failure(t *testing.T) {
	// 必要な環境変数を一時的に設定（DBConfig.Loadで使用する変数名は仮定）
	os.Setenv("DB_USER", "invalid_user")
	os.Setenv("DB_PASSWORD", "invalid_pass")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_DATABASE", "nonexistent_db")

	// テスト終了時に環境変数をリセット
	t.Cleanup(func() {
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_DATABASE")
	})

	conn, err := NewPostgresConnection()
	assert.Error(t, err)
	assert.Nil(t, conn)
}
