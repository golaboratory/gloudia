package realtime

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckOrigin(t *testing.T) {
	t.Run("denies cross-origin when no allowlist is configured (fail-closed)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://api.example.com/ws", nil)
		req.Host = "api.example.com"
		req.Header.Set("Origin", "https://evil.example.com")
		assert.False(t, checkOrigin(req, resolveServeConfig(nil)))
	})

	t.Run("allows same-origin when no allowlist is configured", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://api.example.com/ws", nil)
		req.Host = "api.example.com"
		req.Header.Set("Origin", "https://api.example.com")
		assert.True(t, checkOrigin(req, resolveServeConfig(nil)))
	})

	t.Run("same-origin comparison ignores case", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://api.example.com/ws", nil)
		req.Host = "API.Example.COM"
		req.Header.Set("Origin", "https://api.example.com")
		assert.True(t, checkOrigin(req, resolveServeConfig(nil)))
	})

	t.Run("denies malformed Origin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://api.example.com/ws", nil)
		req.Host = "api.example.com"
		req.Header.Set("Origin", "://not a url")
		assert.False(t, checkOrigin(req, resolveServeConfig(nil)))
	})

	t.Run("enforces allowlist from WS_ALLOWED_ORIGINS", func(t *testing.T) {
		t.Setenv("WS_ALLOWED_ORIGINS", "https://app.example.com, https://admin.example.com")

		denied := httptest.NewRequest("GET", "/ws", nil)
		denied.Header.Set("Origin", "https://evil.example.com")
		assert.False(t, checkOrigin(denied, resolveServeConfig(nil)))

		allowed := httptest.NewRequest("GET", "/ws", nil)
		allowed.Header.Set("Origin", "https://app.example.com")
		assert.True(t, checkOrigin(allowed, resolveServeConfig(nil)))
	})

	t.Run("WithAllowedOrigins takes precedence over the environment variable", func(t *testing.T) {
		t.Setenv("WS_ALLOWED_ORIGINS", "https://from-env.example.com")
		cfg := resolveServeConfig([]ServeOption{WithAllowedOrigins("https://from-option.example.com")})

		fromOption := httptest.NewRequest("GET", "/ws", nil)
		fromOption.Header.Set("Origin", "https://from-option.example.com")
		assert.True(t, checkOrigin(fromOption, cfg))

		fromEnv := httptest.NewRequest("GET", "/ws", nil)
		fromEnv.Header.Set("Origin", "https://from-env.example.com")
		assert.False(t, checkOrigin(fromEnv, cfg))
	})

	t.Run("WithAllowedOrigins trims blanks and ignores empty entries", func(t *testing.T) {
		cfg := resolveServeConfig([]ServeOption{WithAllowedOrigins("  https://app.example.com  ", "", "   ")})
		assert.Equal(t, []string{"https://app.example.com"}, cfg.allowedOrigins)

		req := httptest.NewRequest("GET", "/ws", nil)
		req.Header.Set("Origin", "https://app.example.com")
		assert.True(t, checkOrigin(req, cfg))
	})

	t.Run("WithAnyOrigin allows every origin", func(t *testing.T) {
		cfg := resolveServeConfig([]ServeOption{WithAnyOrigin()})
		req := httptest.NewRequest("GET", "http://api.example.com/ws", nil)
		req.Host = "api.example.com"
		req.Header.Set("Origin", "https://evil.example.com")
		assert.True(t, checkOrigin(req, cfg))
	})

	t.Run("WithAnyOrigin ignores the environment allowlist", func(t *testing.T) {
		t.Setenv("WS_ALLOWED_ORIGINS", "https://app.example.com")
		cfg := resolveServeConfig([]ServeOption{WithAnyOrigin()})
		assert.Empty(t, cfg.allowedOrigins)

		req := httptest.NewRequest("GET", "/ws", nil)
		req.Header.Set("Origin", "https://evil.example.com")
		assert.True(t, checkOrigin(req, cfg))
	})

	t.Run("allows non-browser client without Origin header (allowlist configured)", func(t *testing.T) {
		t.Setenv("WS_ALLOWED_ORIGINS", "https://app.example.com")
		req := httptest.NewRequest("GET", "/ws", nil)
		assert.True(t, checkOrigin(req, resolveServeConfig(nil)))
	})

	t.Run("allows non-browser client without Origin header (fail-closed default)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://api.example.com/ws", nil)
		req.Host = "api.example.com"
		assert.True(t, checkOrigin(req, resolveServeConfig(nil)))
	})
}
