package pdf

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Convert(t *testing.T) {
	// Mock Gotenberg Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify URL
		if r.URL.Path != "/forms/libreoffice/convert" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Verify Method
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Verify Multipart
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("bad multipart"))
			return
		}

		file, _, err := r.FormFile("files")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing files"))
			return
		}
		defer file.Close()

		// Verify Options
		if r.FormValue("landscape") != "true" {
			// This depends on test case options
		}

		// Success response (fake PDF)
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("%PDF-1.4 mock content"))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	t.Run("success conversion", func(t *testing.T) {
		ctx := context.Background()
		src := strings.NewReader("dummy excel content")
		opts := &ConvertOptions{
			Landscape: true,
			Scale:     1.0,
		}

		pdf, err := client.Convert(ctx, "test.xlsx", src, opts)
		assert.NoError(t, err)
		defer pdf.Close()

		content, err := io.ReadAll(pdf)
		assert.NoError(t, err)
		assert.Equal(t, "%PDF-1.4 mock content", string(content))
	})

	t.Run("server error", func(t *testing.T) {
		// Mock error server - Persistent 500
		errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error"))
		}))
		defer errServer.Close()

		clientErr := NewClient(errServer.URL)
		ctx := context.Background()
		src := strings.NewReader("dummy")

		// リトライしても500が続くので最終的にエラーになる
		pdf, err := clientErr.Convert(ctx, "test.xlsx", src, nil)
		assert.Error(t, err)
		assert.Nil(t, pdf)
		// Error should indicate max retries or 500 status (depends on Client impl)
		// For now just checking error existence
	})

	t.Run("retry success", func(t *testing.T) {
		// 最初の2回は500, 3回目で成功するサーバー
		var attempt int32
		retryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&attempt, 1)
			if count < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/pdf")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("%PDF-1.4 retry success"))
		}))
		defer retryServer.Close()

		// Note: httpclient.DefaultConfig() has MaxRetries=3, so it should succeed.
		// テスト時間を短縮したい場合は Config を注入できると良いが、
		// NewClient は引数を取らないためデフォルトのまま実行する。
		// DefaultConfigのRetryWaitMinは1秒なのでテストが数秒かかるが許容する。

		clientRetry := NewClient(retryServer.URL)
		ctx := context.Background()
		src := strings.NewReader("dummy")

		pdf, err := clientRetry.Convert(ctx, "test.xlsx", src, nil)
		assert.NoError(t, err)
		defer pdf.Close()

		content, err := io.ReadAll(pdf)
		assert.NoError(t, err)
		assert.Equal(t, "%PDF-1.4 retry success", string(content))
		assert.Equal(t, int32(3), atomic.LoadInt32(&attempt))
	})
}
