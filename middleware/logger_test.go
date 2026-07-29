package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	// Simple handler that writes something
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	})

	middleware := NewLogger()
	wrappedHandler := middleware(handler)

	t.Run("logs access", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", bytes.NewBufferString("body"))
		rec := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "created", rec.Body.String())
	})

	t.Run("default status is 200", func(t *testing.T) {
		// WriteHeader を呼ばないハンドラ
		noStatusHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})

		mw := NewLogger()
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		mw(noStatusHandler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})

	t.Run("nil body does not panic", func(t *testing.T) {
		mw := NewLogger()
		req := httptest.NewRequest("GET", "/no-body", nil)
		req.Body = nil
		rec := httptest.NewRecorder()

		assert.NotPanics(t, func() {
			mw(handler).ServeHTTP(rec, req)
		})
		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("captures response size", func(t *testing.T) {
		bigHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("hello world"))
		})

		mw := NewLogger()
		req := httptest.NewRequest("GET", "/big", nil)
		rec := httptest.NewRecorder()

		mw(bigHandler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "hello world", rec.Body.String())
	})

	t.Run("debug mode preserves request body for handler", func(t *testing.T) {
		origDebug := IsDebug
		IsDebug = true
		defer func() { IsDebug = origDebug }()

		bodyContent := "request body content"
		var handlerReceivedBody string
		bodyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			handlerReceivedBody = string(b)
			w.WriteHeader(http.StatusOK)
		})

		mw := NewLogger()
		req := httptest.NewRequest("POST", "/debug", bytes.NewBufferString(bodyContent))
		rec := httptest.NewRecorder()

		mw(bodyHandler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, bodyContent, handlerReceivedBody)
	})

	// デバッグモードで 1MB 超のボディが送信された場合、ハンドラにデータが正しく渡ることを確認
	t.Run("debug mode with large body over 1MB", func(t *testing.T) {
		origDebug := IsDebug
		IsDebug = true
		defer func() { IsDebug = origDebug }()

		// 1.5MB のボディを作成
		largeBody := make([]byte, 1536*1024)
		for i := range largeBody {
			largeBody[i] = 'A'
		}

		var handlerReceivedSize int
		bodyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			handlerReceivedSize = len(b)
			w.WriteHeader(http.StatusOK)
		})

		mw := NewLogger()
		req := httptest.NewRequest("POST", "/large", bytes.NewBuffer(largeBody))
		rec := httptest.NewRecorder()

		mw(bodyHandler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		// MultiReader により、ログ読み取り分(1MB) + 残り(0.5MB) がハンドラに渡る
		assert.Equal(t, len(largeBody), handlerReceivedSize)
	})

	t.Run("500 status captured", func(t *testing.T) {
		errHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
		})

		mw := NewLogger()
		req := httptest.NewRequest("GET", "/error", nil)
		rec := httptest.NewRecorder()

		mw(errHandler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "error", rec.Body.String())
	})
}
