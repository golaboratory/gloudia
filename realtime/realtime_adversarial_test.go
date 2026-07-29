package realtime

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtractToken_HeaderQueryPrecedenceAndMalformed は extractToken の
// ヘッダー/クエリの優先順位および不正な Authorization 値の扱いを検証します。
func TestExtractToken_HeaderQueryPrecedenceAndMalformed(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		query      string // raw query string value for token, empty means none
		want       string
	}{
		{name: "header bearer wins over query", authHeader: "Bearer header-token", query: "query-token", want: "header-token"},
		{name: "bearer scheme is case-insensitive", authHeader: "bearer lower-token", query: "", want: "lower-token"},
		{name: "BEARER upper scheme", authHeader: "BEARER up-token", query: "", want: "up-token"},
		{name: "non-bearer scheme falls through to query", authHeader: "Basic abc123", query: "query-token", want: "query-token"},
		{name: "bearer with no value falls through to query", authHeader: "Bearer", query: "query-token", want: "query-token"},
		{name: "empty header uses query", authHeader: "", query: "only-query", want: "only-query"},
		{name: "no token anywhere returns empty", authHeader: "", query: "", want: ""},
		{name: "extra whitespace in header still parses", authHeader: "Bearer    spaced-token", query: "", want: "spaced-token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := "/ws"
			if tt.query != "" {
				target = "/ws?token=" + tt.query
			}
			req := httptest.NewRequest(http.MethodGet, target, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			got := extractToken(req)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestClient_ReadDeadline_TokenExpiryBranches は Client.readDeadline の
// トークン有効期限に基づく3分岐を検証します。
func TestClient_ReadDeadline_TokenExpiryBranches(t *testing.T) {
	t.Run("zero token expiry returns now plus pongWait", func(t *testing.T) {
		c := &Client{} // tokenExpiry is zero value
		before := time.Now()
		got := c.readDeadline()
		// Should be approximately now+pongWait, never the zero time.
		assert.False(t, got.IsZero())
		assert.WithinDuration(t, before.Add(pongWait), got, time.Second)
	})

	t.Run("token expiry sooner than pongWait is used as deadline", func(t *testing.T) {
		expiry := time.Now().Add(5 * time.Second) // well before pongWait (60s)
		c := &Client{tokenExpiry: expiry}
		got := c.readDeadline()
		assert.True(t, got.Equal(expiry), "expected token expiry to be returned, got %v", got)
	})

	t.Run("token expiry later than pongWait falls back to pongWait", func(t *testing.T) {
		expiry := time.Now().Add(pongWait + time.Hour) // far in the future
		c := &Client{tokenExpiry: expiry}
		before := time.Now()
		got := c.readDeadline()
		assert.WithinDuration(t, before.Add(pongWait), got, time.Second)
		assert.True(t, got.Before(expiry), "deadline must not exceed pongWait when token outlives it")
	})
}

// TestHub_Run_BroadcastDropsSlowConsumer は Hub.Run のブロードキャスト処理で
// 送信バッファが満杯の遅いクライアントが切断・除去されることを検証します。
func TestHub_Run_BroadcastDropsSlowConsumer(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// full has a zero-capacity buffer so any send hits the default (drop) branch.
	full := newTenantTestClient(hub, 1, "tenant-a", 0)
	healthy := newTenantTestClient(hub, 2, "tenant-a", 4)
	hub.register <- full
	hub.register <- healthy

	// Wait for both registrations to be reflected in the hub.
	deadline := time.Now().Add(time.Second)
	for {
		hub.mu.RLock()
		n := len(hub.clients)
		hub.mu.RUnlock()
		if n == 2 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("registrations were not reflected")
		}
		time.Sleep(5 * time.Millisecond)
	}

	hub.BroadcastToAll([]byte("msg"))

	// The healthy client must receive the message.
	select {
	case got, ok := <-healthy.send:
		require.True(t, ok, "healthy client channel should remain open")
		assert.Equal(t, "msg", string(got))
	case <-time.After(time.Second):
		t.Fatal("healthy client did not receive broadcast")
	}

	// The slow (full-buffer) client must have been evicted from the hub.
	evicted := false
	deadline = time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		hub.mu.RLock()
		_, present := hub.clients[full]
		n := len(hub.clients)
		hub.mu.RUnlock()
		if !present && n == 1 {
			evicted = true
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	assert.True(t, evicted, "slow consumer should have been removed from hub.clients")

	// Its send channel must have been closed by Run.
	select {
	case _, ok := <-full.send:
		assert.False(t, ok, "evicted client's send channel must be closed")
	default:
		t.Fatal("evicted client's send channel was not closed")
	}
}

// TestHub_BroadcastToTenant_FanoutAndContentIsolation はテナント別配信が
// 同一テナント内の複数クライアントへ正しい内容で届き、かつ同一 userID の
// 別テナントクライアントへは漏えいしないことを検証します。
func TestHub_BroadcastToTenant_FanoutAndContentIsolation(t *testing.T) {
	hub := NewHub()
	// Two clients in tenant-a, including same userID as a tenant-b client to
	// ensure isolation is by tenant, not user.
	a1 := newTenantTestClient(hub, 1, "tenant-a", 4)
	a2 := newTenantTestClient(hub, 2, "tenant-a", 4)
	bSameUser := newTenantTestClient(hub, 1, "tenant-b", 4) // same userID as a1, different tenant
	hub.clients[a1] = true
	hub.clients[a2] = true
	hub.clients[bSameUser] = true

	payload := []byte("tenant-a-only-\xf0\x9f\x94\x92") // includes multibyte content
	hub.BroadcastToTenant("tenant-a", payload)

	// Both tenant-a clients receive the exact payload.
	for i, c := range []*Client{a1, a2} {
		select {
		case got := <-c.send:
			assert.Equal(t, string(payload), string(got), "tenant-a client %d content", i)
		default:
			t.Fatalf("tenant-a client %d did not receive the message", i)
		}
	}

	// The same-userID tenant-b client must NOT receive anything (no cross-tenant leak).
	assert.Equal(t, 0, len(bSameUser.send), "tenant-b client must not receive tenant-a broadcast")
}
