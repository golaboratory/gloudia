package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCORS(t *testing.T) {
	c := NewCORS()
	assert.NotNil(t, c)

	t.Run("returns valid cors handler", func(t *testing.T) {
		// NewCORS が nil でないハンドラを返すことを検証
		handler := c.Handler(nil)
		assert.NotNil(t, handler)
	})
}

func TestNewCORS_OriginAllowlist(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("arbitrary https origin is NOT reflected", func(t *testing.T) {
		handler := NewCORS().Handler(next)
		req := httptest.NewRequest("GET", "http://api.example.com/resource", nil)
		req.Header.Set("Origin", "https://evil.example.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// 任意の HTTPS オリジンが Access-Control-Allow-Origin に反射されないこと
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("origin configured via CORS_ALLOWED_ORIGINS is allowed", func(t *testing.T) {
		t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com")
		handler := NewCORS().Handler(next)
		req := httptest.NewRequest("GET", "http://api.example.com/resource", nil)
		req.Header.Set("Origin", "https://app.example.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
	})
}
