package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisRateLimiter(t *testing.T) {
	// Setup miniredis
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := RateLimitConfig{
		Rate:   2,
		Burst:  2, // Allow 2 requests immediately
		Period: time.Minute,
		Name:   "test_limit",
	}

	limiterMiddleware := NewRedisRateLimiter(rdb, config)

	t.Run("allow requests", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			r.Header.Set("X-Real-IP", "127.0.0.1")
			ctx := humatest.NewContext(nil, r, w)

			nextCalled := false
			limiterMiddleware(ctx, func(c huma.Context) {
				nextCalled = true
			})

			assert.True(t, nextCalled, "Request %d should be allowed", i)
			assert.Equal(t, strconv.Itoa(config.Rate), w.Result().Header.Get("X-RateLimit-Limit"))
		}
	})

	t.Run("block request", func(t *testing.T) {
		// 3rd request should be blocked
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("X-Real-IP", "127.0.0.1") // Same IP
		ctx := humatest.NewContext(nil, r, w)

		nextCalled := false
		limiterMiddleware(ctx, func(c huma.Context) {
			nextCalled = true
		})

		assert.False(t, nextCalled)
		assert.Equal(t, http.StatusTooManyRequests, w.Result().StatusCode)

		var body map[string]interface{}
		json.NewDecoder(w.Body).Decode(&body)
		assert.Equal(t, float64(429), body["status"])
	})

	t.Run("fail open when redis down", func(t *testing.T) {
		mr.Close() // Simulate redis down

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("X-Real-IP", "127.0.0.2") // New IP
		ctx := humatest.NewContext(nil, r, w)

		nextCalled := false
		limiterMiddleware(ctx, func(c huma.Context) {
			nextCalled = true
		})

		// Should proceed despite error
		assert.True(t, nextCalled)
	})
}
