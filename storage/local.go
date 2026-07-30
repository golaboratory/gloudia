package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/newmo-oss/ergo"
)

var (
	// ErrInvalidPath は指定されたパスがベースディレクトリの外を指す場合
	// （パストラバーサル）や、パスの解決に失敗した場合のエラーです。
	ErrInvalidPath = ergo.NewSentinel("path traversal detected: path escapes base directory")
)

// ctxReader は読み取りごとに context のキャンセルを確認する io.Reader ラッパーです。
// キャンセル/タイムアウト時にコピーを速やかに中断し、ファイルハンドル等のリソースを解放します。
type ctxReader struct {
	ctx context.Context
	r   io.Reader
}

func (cr *ctxReader) Read(p []byte) (int, error) {
	if err := cr.ctx.Err(); err != nil {
		return 0, err
	}
	return cr.r.Read(p)
}

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

func (s *LocalStorage) getFullPath(path string) (string, error) {
	// パスを結合してから正規化し、baseDir 配下に収まることを検証する
	joined := filepath.Join(s.baseDir, path)
	absPath, err := filepath.Abs(joined)
	if err != nil {
		return "", ergo.Wrap(err, "failed to resolve absolute path")
	}
	absBase, err := filepath.Abs(s.baseDir)
	if err != nil {
		return "", ergo.Wrap(err, "failed to resolve base directory")
	}

	// シンボリックリンクを解決してから前方一致で判定する。
	// filepath.Abs は字句的な正規化 (.. の解決) のみでシンボリックリンクを辿らないため、
	// baseDir 内に外部を指すシンボリックリンクが存在すると、それを経由して
	// baseDir 外のファイルを読み書き/削除できてしまう (パストラバーサル)。
	resolvedBase, err := filepath.EvalSymlinks(absBase)
	if err != nil {
		return "", ergo.Wrap(err, "failed to resolve base directory symlinks")
	}
	// 対象パスは未作成 (Upload 先) の場合があるため、存在する最も近い祖先まで遡って解決する。
	resolvedPath, err := evalSymlinksAllowingMissing(absPath)
	if err != nil {
		return "", ergo.Wrap(err, "failed to resolve path symlinks")
	}

	// baseDir 配下であることを確認 (separator を付与して前方一致で判定)
	if !strings.HasPrefix(resolvedPath, resolvedBase+string(filepath.Separator)) && resolvedPath != resolvedBase {
		return "", ErrInvalidPath
	}
	// 検証対象 (resolvedPath) と実際の操作対象を一致させ、字句パス(absPath)経由での
	// シンボリックリンク追従による baseDir 外への脱出を防ぐ。
	return resolvedPath, nil
}

// evalSymlinksAllowingMissing は filepath.EvalSymlinks と同様にシンボリックリンクを
// 解決しますが、パスがまだ存在しない場合 (Upload 先など) は、存在する最も近い祖先
// ディレクトリまで遡って解決し、残りの未存在要素を結合して返します。
// これにより、未作成パスでもシンボリックリンクによる baseDir 外への脱出を検出できます。
//
// 重要: 最終要素がダングリング・シンボリックリンク (リンク先が未作成) の場合、
// EvalSymlinks は ENOENT を返すが、その最終リンクを字句的に再結合してしまうと
// リンク先 (baseDir 外) を解決できず脱出を許す。そのため Lstat で最終要素が
// シンボリックリンクかを確認し、シンボリックリンクならリンク先を辿って解決する。
func evalSymlinksAllowingMissing(p string) (string, error) {
	resolved, err := filepath.EvalSymlinks(p)
	if err == nil {
		return resolved, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}

	// p 自体が (リンク先未作成の) シンボリックリンクの可能性があるため確認する。
	if fi, lerr := os.Lstat(p); lerr == nil && fi.Mode()&os.ModeSymlink != 0 {
		target, rerr := os.Readlink(p)
		if rerr != nil {
			return "", rerr
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(p), target)
		}
		// 解決済みパスへの無限再帰を避けるため Clean してから辿る。
		return evalSymlinksAllowingMissing(filepath.Clean(target))
	}

	// p はシンボリックリンクとしても存在しない純粋な新規要素。
	// 存在する最も近い祖先まで遡って解決し、残りの未存在要素を字句的に結合する。
	parent := filepath.Dir(p)
	if parent == p {
		// ルートに到達 (これ以上遡れない)
		return p, nil
	}
	resolvedParent, perr := evalSymlinksAllowingMissing(parent)
	if perr != nil {
		return "", perr
	}
	return filepath.Join(resolvedParent, filepath.Base(p)), nil
}

// Upload は指定されたパスにデータをアップロードします。
// baseDir の外を指すパス（パストラバーサル）は ErrInvalidPath で拒否します。
func (s *LocalStorage) Upload(ctx context.Context, path string, data io.Reader) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	fullPath, err := s.getFullPath(path)
	if err != nil {
		return err
	}

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

	// context のキャンセルを尊重しながらコピーする
	if _, err := io.Copy(file, &ctxReader{ctx: ctx, r: data}); err != nil {
		return ergo.Wrap(err, "failed to write data")
	}

	return nil
}

// Download は指定されたパスのデータをダウンロードするためのReaderを返します。
// 呼び出し元はReadCloserをCloseする責任があります。
// baseDir の外を指すパス（パストラバーサル）は ErrInvalidPath で拒否します。
func (s *LocalStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	fullPath, err := s.getFullPath(path)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, ergo.Wrap(err, "failed to open file")
	}

	return file, nil
}

// Delete は指定されたパスのファイルを削除します。
// ファイルが既に存在しない場合は成功とみなし nil を返します。
// baseDir の外を指すパス（パストラバーサル）は ErrInvalidPath で拒否します。
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	fullPath, err := s.getFullPath(path)
	if err != nil {
		return err
	}

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
// 実際の署名検証は行われない疑似的なものです（method と expires 引数は無視されます）。
// baseDir の外を指すパス（パストラバーサル）は ErrInvalidPath で拒否します。
func (s *LocalStorage) GetSignedURL(ctx context.Context, path string, method string, expires time.Duration) (string, error) {
	// 他のメソッド（Upload/Download/Delete）と同様にパストラバーサル検証を行い、
	// baseDir 外を指すパスから URL を生成しないようにする。
	if _, err := s.getFullPath(path); err != nil {
		return "", err
	}
	// WindowsのパスセパレータをURL用にスラッシュに変換
	urlPath := filepath.ToSlash(path)
	return fmt.Sprintf("%s/%s", s.baseURL, urlPath), nil
}
