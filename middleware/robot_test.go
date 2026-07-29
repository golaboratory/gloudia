package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRobotTag(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewRobotTag()
	wrappedHandler := middleware(handler)

	t.Run("sets X-Robots-Tag header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "noindex, nofollow, noarchive", rec.Header().Get("X-Robots-Tag"))
	})

	t.Run("calls next handler", func(t *testing.T) {
		nextCalled := false
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.Write([]byte("ok"))
		})

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		middleware(inner).ServeHTTP(rec, req)

		assert.True(t, nextCalled)
		assert.Equal(t, "ok", rec.Body.String())
	})
}
