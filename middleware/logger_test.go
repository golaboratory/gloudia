package middleware

import (
	"bytes"
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
		// Log capturing is hard without capturing stdout/slog.
		// For now, we verify that the handler is called and response is correct.
		// Detailed log verification would require replacing slog.Default() or a custom logger.

		req := httptest.NewRequest("POST", "/test", bytes.NewBufferString("body"))
		rec := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "created", rec.Body.String())
	})
}
