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

	t.Run("Upload rejects path traversal", func(t *testing.T) {
		reader := strings.NewReader("malicious")
		err := storage.Upload(ctx, "../../../etc/passwd", reader)
		assert.ErrorIs(t, err, ErrInvalidPath)
	})

	t.Run("Download rejects path traversal", func(t *testing.T) {
		_, err := storage.Download(ctx, "../../secret.txt")
		assert.ErrorIs(t, err, ErrInvalidPath)
	})

	t.Run("Delete rejects path traversal", func(t *testing.T) {
		err := storage.Delete(ctx, "../outside.txt")
		assert.ErrorIs(t, err, ErrInvalidPath)
	})

	t.Run("Upload allows nested subdirectory", func(t *testing.T) {
		reader := strings.NewReader("nested content")
		err := storage.Upload(ctx, "a/b/c/file.txt", reader)
		assert.NoError(t, err)
	})

	t.Run("rejects traversal via symlink escaping baseDir", func(t *testing.T) {
		// baseDir 外を指すシンボリックリンクを baseDir 内に作成する
		outsideDir, err := os.MkdirTemp("", "storage_outside")
		assert.NoError(t, err)
		defer os.RemoveAll(outsideDir)

		secret := filepath.Join(outsideDir, "secret.txt")
		assert.NoError(t, os.WriteFile(secret, []byte("top secret"), 0600))

		linkPath := filepath.Join(tempDir, "escape")
		if err := os.Symlink(outsideDir, linkPath); err != nil {
			t.Skipf("symlink not supported on this platform: %v", err)
		}

		// "escape/secret.txt" は字句的には baseDir 配下だが、実体は baseDir 外を指す
		_, err = storage.Download(ctx, "escape/secret.txt")
		assert.ErrorIs(t, err, ErrInvalidPath)

		// 未作成パスでもシンボリックリンク経由の脱出を検出する
		err = storage.Upload(ctx, "escape/evil.txt", strings.NewReader("x"))
		assert.ErrorIs(t, err, ErrInvalidPath)

		err = storage.Delete(ctx, "escape/secret.txt")
		assert.ErrorIs(t, err, ErrInvalidPath)
	})

	t.Run("rejects dangling final-component symlink (arbitrary write)", func(t *testing.T) {
		// 最終要素自体が「リンク先未作成」のシンボリックリンクのケース。
		// 字句的な再結合では検出できず、os.Create がリンクを追従して
		// baseDir 外へ任意書き込みが可能になる脆弱性の回帰テスト。
		outsideDir, err := os.MkdirTemp("", "storage_dangling")
		assert.NoError(t, err)
		defer os.RemoveAll(outsideDir)

		ghost := filepath.Join(outsideDir, "ghost.txt") // 存在しないリンク先
		linkPath := filepath.Join(tempDir, "plant")
		if err := os.Symlink(ghost, linkPath); err != nil {
			t.Skipf("symlink not supported on this platform: %v", err)
		}

		// Upload が拒否され、baseDir 外へ書き込まれていないことを確認
		err = storage.Upload(ctx, "plant", strings.NewReader("PWNED"))
		assert.ErrorIs(t, err, ErrInvalidPath)
		_, statErr := os.Stat(ghost)
		assert.True(t, os.IsNotExist(statErr), "must not have written outside baseDir")
	})
}
