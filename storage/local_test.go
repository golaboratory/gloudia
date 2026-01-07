package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLocalStorage(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "storage_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir) // テスト終了後にクリーンアップ

	storage, err := NewLocalStorage(tempDir, "http://localhost:8080/static")
	assert.NoError(t, err)

	ctx := context.Background()
	testFile := "test_folder/sample.txt"
	content := "Hello, Storage!"

	t.Run("Upload", func(t *testing.T) {
		reader := strings.NewReader(content)
		err := storage.Upload(ctx, testFile, reader)
		assert.NoError(t, err)

		// 実際にファイルが作成されたか確認
		fullPath := filepath.Join(tempDir, testFile)
		_, err = os.Stat(fullPath)
		assert.NoError(t, err)
	})

	t.Run("Download", func(t *testing.T) {
		rc, err := storage.Download(ctx, testFile)
		assert.NoError(t, err)
		defer rc.Close()

		data, err := io.ReadAll(rc)
		assert.NoError(t, err)
		assert.Equal(t, content, string(data))
	})

	t.Run("GetSignedURL", func(t *testing.T) {
		url, err := storage.GetSignedURL(ctx, testFile, "GET", time.Hour)
		assert.NoError(t, err)
		// URLセパレータがスラッシュであることを確認
		expected := "http://localhost:8080/static/test_folder/sample.txt"
		assert.Equal(t, expected, url)
	})

	t.Run("Delete", func(t *testing.T) {
		err := storage.Delete(ctx, testFile)
		assert.NoError(t, err)

		// ファイルが削除されたか確認
		fullPath := filepath.Join(tempDir, testFile)
		_, err = os.Stat(fullPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("Delete Non-existent", func(t *testing.T) {
		// 存在しないファイルを削除してもエラーにならないこと
		err := storage.Delete(ctx, "non_existent.txt")
		assert.NoError(t, err)
	})
}
