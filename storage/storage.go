package storage

import (
	"context"
	"io"
	"time"
)

// Storage はファイルストレージへの操作を抽象化するインターフェースです。
type Storage interface {
	// Upload は指定されたパスにデータをアップロードします。
	Upload(ctx context.Context, path string, data io.Reader) error

	// Download は指定されたパスのデータをダウンロードするためのReaderを返します。
	// 呼び出し元はReadCloserをCloseする責任があります。
	Download(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete は指定されたパスのファイルを削除します。
	Delete(ctx context.Context, path string) error

	// GetSignedURL は指定されたパスへの署名付きURL（期限付きアクセスURL）を発行します。
	// method は "GET" (ダウンロード用) や "PUT" (アップロード用) などを指定します。
	GetSignedURL(ctx context.Context, path string, method string, expires time.Duration) (string, error)
}
