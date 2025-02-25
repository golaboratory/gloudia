package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadEnvironmentsWithFile_NonExistent(t *testing.T) {
	nonexistentPath := "not_exist.env"
	err := LoadEnvironmentsWithFile(nonexistentPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "環境ファイルが存在しません")
}

func TestLoadEnvironmentsWithFile_Success(t *testing.T) {
	// 一時ディレクトリにテスト用の.envファイルを作成
	tempDir := t.TempDir()
	testEnvPath := filepath.Join(tempDir, ".env")
	content := []byte("TEST_VAR=testvalue\n")
	err := os.WriteFile(testEnvPath, content, 0644)
	assert.NoError(t, err)

	// 存在する.envファイルをロード
	err = LoadEnvironmentsWithFile(testEnvPath)
	assert.NoError(t, err)

	// 環境変数が正しく読み込まれているか検証
	assert.Equal(t, "testvalue", os.Getenv("TEST_VAR"))

	// テスト終了後に環境変数をクリーンアップ
	os.Unsetenv("TEST_VAR")
}
