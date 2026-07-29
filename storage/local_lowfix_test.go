package storage

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetSignedURL_RejectsTraversal(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_signed")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	s, err := NewLocalStorage(tempDir, "http://localhost/static")
	assert.NoError(t, err)

	// GetSignedURL も他メソッドと同様にトラバーサルを拒否すること
	_, err = s.GetSignedURL(context.Background(), "../../etc/passwd", "GET", time.Hour)
	assert.ErrorIs(t, err, ErrInvalidPath)
}

func TestLocalStorage_HonorsContextCancellation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_ctx")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	s, err := NewLocalStorage(tempDir, "http://localhost/static")
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 即時キャンセル

	assert.ErrorIs(t, s.Upload(ctx, "f.txt", strings.NewReader("x")), context.Canceled)
	_, dErr := s.Download(ctx, "f.txt")
	assert.ErrorIs(t, dErr, context.Canceled)
	assert.ErrorIs(t, s.Delete(ctx, "f.txt"), context.Canceled)
}
