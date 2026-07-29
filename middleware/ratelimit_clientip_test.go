package middleware

import (
	"net"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimitClientIP(t *testing.T) {
	newCtx := func(remoteAddr, xRealIP string) huma.Context {
		r := httptest.NewRequest("GET", "/test", nil)
		r.RemoteAddr = remoteAddr
		if xRealIP != "" {
			r.Header.Set("X-Real-IP", xRealIP)
		}
		return humatest.NewContext(nil, r, httptest.NewRecorder())
	}

	_, cidr10, err := net.ParseCIDR("10.0.0.0/8")
	require.NoError(t, err)
	trusted := []*net.IPNet{cidr10}

	t.Run("untrusted (public) peer: spoofable X-Real-IP is ignored, keys on direct peer", func(t *testing.T) {
		// 攻撃者がパブリック経路から直接接続し X-Real-IP を偽装しても無視されること
		ctx := newCtx("203.0.113.9:5000", "10.1.2.3")
		assert.Equal(t, "203.0.113.9", rateLimitClientIP(ctx, trusted))
	})

	t.Run("explicitly trusted peer: honors X-Real-IP", func(t *testing.T) {
		// TRUSTED_PROXY_CIDRS で明示した範囲の送信元のみ X-Real-IP を採用する
		ctx := newCtx("10.0.0.5:5000", "198.51.100.7")
		assert.Equal(t, "198.51.100.7", rateLimitClientIP(ctx, trusted))
	})

	t.Run("no trusted CIDR configured: private peer's X-Real-IP is ignored (fail-closed)", func(t *testing.T) {
		// 許可リスト未設定時はプライベートアドレスからのヘッダーも信頼しない
		ctx := newCtx("10.0.0.5:5000", "198.51.100.7")
		assert.Equal(t, "10.0.0.5", rateLimitClientIP(ctx, nil))
	})
}
