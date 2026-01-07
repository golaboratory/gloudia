package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/golaboratory/gloudia/auth"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthProvider(t *testing.T) {
	// Setup TokenMaker
	hexKey := auth.GenerateRandomKey()
	maker, err := auth.NewTokenMaker(hexKey)
	assert.NoError(t, err)

	provider := NewAuthProvider(maker)

	t.Run("no header", func(t *testing.T) {
		// Create a test context
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		ctx := humatest.NewContext(nil, r, w)

		// Call middleware
		nextCalled := false
		provider(ctx, func(c huma.Context) {
			nextCalled = true
		})

		// Should skip verify and call next?
		// Code says: if authHeader == "" { ctx.SetStatus(http.StatusUnauthorized); return }
		// Wait, line 28 in auth.go:
		// if authHeader == "" { ctx.SetStatus(http.StatusUnauthorized); return }
		// So it blocks guests. The comments said "verification skip... (guest access consideration)" but the CODE does return 401.
		// Actually, comment says "ヘッダーがない場合は検証をスキップして次の処理へ" (Skip verification and proceed).
		// BUT the code:
		// if authHeader == "" {
		//	ctx.SetStatus(http.StatusUnauthorized)
		//	return
		//}
		// This contradicts the comment?
		// Wait, looking at file view in Step 65:
		// 26: 		// ヘッダーがない場合は検証をスキップして次の処理へ (ゲストアクセスの可能性)
		// 28: 		if authHeader == "" {
		// 29: 			ctx.SetStatus(http.StatusUnauthorized)
		// 30: 			return
		// 31: 		}
		// The code definitely returns 401. The comment might be stale or logic changed.
		// I will assert based on Code. Code returns 401.

		assert.False(t, nextCalled)
		assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})

	t.Run("invalid header format", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("Authorization", "Basic user:pass")
		ctx := humatest.NewContext(nil, r, w)

		nextCalled := false
		provider(ctx, func(c huma.Context) {
			nextCalled = true
		})

		assert.False(t, nextCalled)
		assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})

	t.Run("valid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		// Create token
		token, err := maker.CreateToken(1, "tenant1", 2, time.Minute)
		assert.NoError(t, err)

		r.Header.Set("Authorization", "Bearer "+token)
		ctx := humatest.NewContext(nil, r, w)

		nextCalled := false
		provider(ctx, func(c huma.Context) {
			nextCalled = true
			// Verify claims are in context
			claims, ok := c.Context().Value(KeyClaims).(*auth.Claims)
			assert.True(t, ok)
			assert.Equal(t, int64(1), claims.UserID)
			assert.Equal(t, "tenant1", claims.TenantID)
		})

		assert.True(t, nextCalled)
	})

	t.Run("expired token", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		// Create token with negative duration
		token, err := maker.CreateToken(1, "tenant1", 2, -time.Minute)
		assert.NoError(t, err)

		r.Header.Set("Authorization", "Bearer "+token)
		ctx := humatest.NewContext(nil, r, w)

		nextCalled := false
		provider(ctx, func(c huma.Context) {
			nextCalled = true
		})

		assert.False(t, nextCalled)
		assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})
}
