package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/newmo-oss/ergo"
)

// LocalStorage はローカルファイルシステムを使用するストレージ実装です。
// 開発環境や、永続化ボリュームがマウントされた環境での使用を想定しています。
type LocalStorage struct {
	baseDir string
	baseURL string // GetSignedURLで返す際のベースURL (例: http://localhost:8080/files)
}

// NewLocalStorage は新しいLocalStorageインスタンスを作成します。
// baseDir: ファイルを保存するルートディレクトリ
// baseURL: 署名付きURL(模倣)のプレフィックス
func NewLocalStorage(baseDir string, baseURL string) (*LocalStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, ergo.Wrap(err, "failed to create base directory")
	}
	return &LocalStorage{
		baseDir: baseDir,
		baseURL: baseURL,
	}, nil
}

func (s *LocalStorage) getFullPath(path string) string {
	// ディレクトリトラバーサル対策として、ファイル名のみを結合するなどが本来必要ですが、
	// ここでは単純に結合します。
	return filepath.Join(s.baseDir, path)
}

func (s *LocalStorage) Upload(ctx context.Context, path string, data io.Reader) error {
	fullPath := s.getFullPath(path)

	// 親ディレクトリの作成
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ergo.Wrap(err, "failed to create directory")
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return ergo.Wrap(err, "failed to create file")
	}
	defer file.Close()

	if _, err := io.Copy(file, data); err != nil {
		return ergo.Wrap(err, "failed to write data")
	}

	return nil
}

func (s *LocalStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := s.getFullPath(path)

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, ergo.Wrap(err, "failed to open file")
	}

	return file, nil
}

func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := s.getFullPath(path)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			// 既に存在しない場合は成功とみなす
			return nil
		}
		return ergo.Wrap(err, "failed to delete file")
	}
	return nil
}

// GetSignedURL はローカル環境においては署名機能を持たないため、
// 単純に静的ファイル配信サーバーへのパスを返します。
// 実際の署名検証は行われない疑似的なものです。
func (s *LocalStorage) GetSignedURL(ctx context.Context, path string, method string, expires time.Duration) (string, error) {
	// WindowsのパスセパレータをURL用にスラッシュに変換
	urlPath := filepath.ToSlash(path)
	return fmt.Sprintf("%s/%s", s.baseURL, urlPath), nil
}
