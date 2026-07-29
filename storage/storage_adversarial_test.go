package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- SakuraObjectStorage ---

// TestSakuraObjectStorage_GetSignedURL_UnsupportedMethod は、GET/PUT/DELETE 以外の
// メソッドが指定された場合に GetSignedURL がネットワークアクセス前に
// ErrUnsupportedMethod を返すことを検証します (オフラインで完結)。
func TestSakuraObjectStorage_GetSignedURL_UnsupportedMethod(t *testing.T) {
	ctx := context.Background()
	cfg := SakuraStorageConfig{
		AccessKey: "dummy_access",
		SecretKey: "dummy_secret",
		Endpoint:  "https://s3.isk01.sakurastorage.jp",
		Bucket:    "my-bucket",
	}
	s, err := NewSakuraObjectStorage(ctx, cfg)
	require.NoError(t, err)

	for _, method := range []string{"PATCH", "HEAD", "get", "post", ""} {
		t.Run("method="+method, func(t *testing.T) {
			url, err := s.GetSignedURL(ctx, "some/key.txt", method, time.Hour)
			assert.ErrorIs(t, err, ErrUnsupportedMethod)
			assert.Contains(t, err.Error(), method)
			assert.Empty(t, url)
		})
	}
}

// TestDefaultSakuraStorageConfig は DefaultSakuraStorageConfig が
// ドキュメント化された既定のエンドポイント/リージョンを返し、
// 認証情報・バケットは呼び出し側が埋める前提で空であることを検証します。
func TestDefaultSakuraStorageConfig(t *testing.T) {
	cfg := DefaultSakuraStorageConfig()
	assert.Equal(t, "https://s3.isk01.sakurastorage.jp", cfg.Endpoint)
	assert.Equal(t, "jp-north-1", cfg.Region)
	// 認証情報とバケットは呼び出し側が埋める前提で空であること
	assert.Empty(t, cfg.AccessKey)
	assert.Empty(t, cfg.SecretKey)
	assert.Empty(t, cfg.Bucket)
}

// --- LocalStorage ---

// TestLocalStorage_UploadOverwriteTruncates は、より短いペイロードで既存ファイルを
// 上書きした際に古い末尾バイトが残らない (os.Create による切り詰め) ことを検証します。
func TestLocalStorage_UploadOverwriteTruncates(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_overwrite")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	s, err := NewLocalStorage(tempDir, "http://localhost/static")
	require.NoError(t, err)
	ctx := context.Background()

	require.NoError(t, s.Upload(ctx, "f.txt", strings.NewReader("longer original content")))
	require.NoError(t, s.Upload(ctx, "f.txt", strings.NewReader("new")))

	rc, err := s.Download(ctx, "f.txt")
	require.NoError(t, err)
	defer rc.Close()
	data, err := io.ReadAll(rc)
	require.NoError(t, err)
	// 切り詰められ、古い内容が残っていないこと
	assert.Equal(t, "new", string(data))
}

// TestLocalStorage_DownloadNonExistentReturnsNonTraversalError は、存在しないが
// 境界内の正当なパスの Download が、トラバーサル拒否 (ErrInvalidPath) とは
// 区別可能な通常の不在エラーを返すことを検証します。
func TestLocalStorage_DownloadNonExistentReturnsNonTraversalError(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_dlmissing")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	s, err := NewLocalStorage(tempDir, "http://localhost/static")
	require.NoError(t, err)

	rc, err := s.Download(context.Background(), "does/not/exist.txt")
	assert.Error(t, err)
	assert.Nil(t, rc)
	// 通常の不在エラーはトラバーサル拒否と区別されること
	assert.NotErrorIs(t, err, ErrInvalidPath)
}

// TestLocalStorage_GetSignedURL_PathEncoding は GetSignedURL のパス処理を固定します。
// filepath.ToSlash によりバックスラッシュはスラッシュへ変換されるが、
// パーセントエンコードは行われず、空白や多バイト文字はそのまま透過することを検証します。
func TestLocalStorage_GetSignedURL_PathEncoding(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_urlenc")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	s, err := NewLocalStorage(tempDir, "http://localhost/files")
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("multibyte and space pass through verbatim (no percent-encoding)", func(t *testing.T) {
		url, err := s.GetSignedURL(ctx, "folder/a b/日本語.txt", "GET", time.Hour)
		require.NoError(t, err)
		assert.Equal(t, "http://localhost/files/folder/a b/日本語.txt", url)
	})

	t.Run("backslash separators are converted to forward slash", func(t *testing.T) {
		url, err := s.GetSignedURL(ctx, filepath.FromSlash("a/b/c.txt"), "GET", time.Hour)
		require.NoError(t, err)
		assert.Equal(t, "http://localhost/files/a/b/c.txt", url)
	})
}

// TestLocalStorage_ConcurrentUploads は、同一の深いネストディレクトリへ並行に
// Upload しても (MkdirAll/Create の競合) パニック・エラー・データ競合が
// 発生しないことを検証します (-race 下での実行を想定)。
func TestLocalStorage_ConcurrentUploads(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_concurrent")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	s, err := NewLocalStorage(tempDir, "http://localhost/static")
	require.NoError(t, err)
	ctx := context.Background()

	const n = 24
	errs := make(chan error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// 同一の深いディレクトリへ並行に書き込む (MkdirAll/Create の競合)
			path := fmt.Sprintf("shared/nested/dir/file_%d.txt", idx)
			errs <- s.Upload(ctx, path, strings.NewReader("payload"))
		}(i)
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		assert.NoError(t, e)
	}

	// 全ファイルが実在し読み出せること
	for i := 0; i < n; i++ {
		rc, err := s.Download(ctx, fmt.Sprintf("shared/nested/dir/file_%d.txt", i))
		require.NoError(t, err)
		data, err := io.ReadAll(rc)
		rc.Close()
		require.NoError(t, err)
		assert.Equal(t, "payload", string(data))
	}
}
