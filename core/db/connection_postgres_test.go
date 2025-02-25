package db

import (
	"context"
	"os"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
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

// TestNewPostgresConnection_Success は、embedded-postgresを使用してPostgresへの接続が成功することを検証します。
func TestNewPostgresConnection_Success(t *testing.T) {
	// 接続成功用の環境変数を設定
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_DATABASE", "postgres")
	t.Cleanup(func() {
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_DATABASE")
	})

	epg := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Username("postgres").
		Password("postgres").
		Database("postgres").
		Version(embeddedpostgres.V12).
		Port(5433))
	err := epg.Start()
	if err != nil {
		t.Fatalf("failed to start embedded-postgres: %v", err)
	}
	// テスト後にembedded-postgresを停止
	defer func() {
		if err := epg.Stop(); err != nil {
			t.Log("failed to stop embedded-postgres:", err)
		}
	}()

	conn, err := NewPostgresConnection()
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	// pgx.Conn のPingで接続を確認
	err = conn.Ping(context.Background())
	assert.NoError(t, err)

	_ = conn.Close(context.Background())
}
