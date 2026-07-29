package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_Do(t *testing.T) {
	t.Run("success without retry", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		conf := DefaultConfig()
		client := NewClient(conf)

		req, _ := http.NewRequest("GET", server.URL, nil)
		resp, err := client.Do(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("success after retry", func(t *testing.T) {
		var accessCount int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&accessCount, 1)
			if count < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		conf := DefaultConfig()
		conf.RetryWaitMin = 1 * time.Millisecond // fast test
		conf.RetryWaitMax = 10 * time.Millisecond
		client := NewClient(conf)

		req, _ := http.NewRequest("GET", server.URL, nil)
		resp, err := client.Do(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, int32(3), atomic.LoadInt32(&accessCount))
	})

	t.Run("failure after max retries", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		conf := DefaultConfig()
		conf.MaxRetries = 2
		conf.RetryWaitMin = 1 * time.Millisecond
		client := NewClient(conf)

		req, _ := http.NewRequest("GET", server.URL, nil)
		resp, err := client.Do(req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "max retries reached")
	})

	t.Run("client timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond) // IDLE
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		conf := DefaultConfig()
		conf.Timeout = 50 * time.Millisecond // Timeout before response
		client := NewClient(conf)

		req, _ := http.NewRequest("GET", server.URL, nil)
		resp, err := client.Do(req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		// Error message varies by OS/Lib but should be related to timeout
	})

	// 全リトライが 500 応答の場合、err が nil でも正しくエラーを返すことを確認
	t.Run("all retries return 500 with nil network error", func(t *testing.T) {
		var accessCount int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&accessCount, 1)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		conf := DefaultConfig()
		conf.MaxRetries = 1
		conf.RetryWaitMin = 1 * time.Millisecond
		conf.RetryWaitMax = 2 * time.Millisecond
		client := NewClient(conf)

		req, _ := http.NewRequest("GET", server.URL, nil)
		resp, err := client.Do(req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "max retries reached")
		// ネットワークエラーなしで 500 のみの場合、2回アクセスされることを確認
		assert.Equal(t, int32(2), atomic.LoadInt32(&accessCount))
	})

	t.Run("context cancel", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(500 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		conf := DefaultConfig()
		client := NewClient(conf)

		ctx, cancel := context.WithCancel(context.Background())
		req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)

		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		resp, err := client.Do(req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, context.Canceled)
	})
}
