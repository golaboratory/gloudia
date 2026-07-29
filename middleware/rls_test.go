package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/assert"
)

func TestNewRLSProvider_NoTenantID(t *testing.T) {
	// db=nil でも tenant_id が空の段階で 400 を返すため pgxpool は不要
	provider := NewRLSProvider(nil)

	t.Run("returns 400 when tenant_id is missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		ctx := humatest.NewContext(nil, r, w)

		nextCalled := false
		provider(ctx, func(c huma.Context) {
			nextCalled = true
		})

		assert.False(t, nextCalled)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("returns 400 when tenant_id is not UUID format", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		// Context に不正な tenant_id をセット
		rCtx := context.WithValue(r.Context(), KeyTenantID, "invalid-tenant'; DROP TABLE users;--")
		r = r.WithContext(rCtx)
		ctx := humatest.NewContext(nil, r, w)

		nextCalled := false
		provider(ctx, func(c huma.Context) {
			nextCalled = true
		})

		assert.False(t, nextCalled)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("returns 400 when tenant_id is empty string in context", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		rCtx := context.WithValue(r.Context(), KeyTenantID, "")
		r = r.WithContext(rCtx)
		ctx := humatest.NewContext(nil, r, w)

		nextCalled := false
		provider(ctx, func(c huma.Context) {
			nextCalled = true
		})

		assert.False(t, nextCalled)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})
}

func TestUUIDPattern(t *testing.T) {
	t.Run("valid UUIDs", func(t *testing.T) {
		validUUIDs := []string{
			"550e8400-e29b-41d4-a716-446655440000",
			"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			"f47ac10b-58cc-4372-a567-0e02b2c3d479",
			"00000000-0000-0000-0000-000000000000",
		}
		for _, id := range validUUIDs {
			assert.True(t, uuidPattern.MatchString(id), "should match: %s", id)
		}
	})

	t.Run("invalid UUIDs", func(t *testing.T) {
		invalidUUIDs := []string{
			"",
			"not-a-uuid",
			"550e8400e29b41d4a716446655440000",      // ハイフンなし
			"550e8400-e29b-41d4-a716-44665544000",   // 短すぎ
			"550e8400-e29b-41d4-a716-4466554400000", // 長すぎ
			"'; DROP TABLE users;--",                // SQLインジェクション
			"550e8400-e29b-41d4-a716-44665544000g",  // 無効な文字
		}
		for _, id := range invalidUUIDs {
			assert.False(t, uuidPattern.MatchString(id), "should not match: %s", id)
		}
	})
}
