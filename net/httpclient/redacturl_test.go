package httpclient

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactURL(t *testing.T) {
	parse := func(t *testing.T, raw string) *url.URL {
		t.Helper()
		u, err := url.Parse(raw)
		require.NoError(t, err)
		return u
	}

	t.Run("nil URL returns empty string", func(t *testing.T) {
		assert.Equal(t, "", redactURL(nil, false))
	})

	t.Run("plain URL is unchanged", func(t *testing.T) {
		u := parse(t, "https://api.example.com/v1/users?page=2")
		assert.Equal(t, "https://api.example.com/v1/users?page=2", redactURL(u, false))
	})

	t.Run("userinfo is stripped", func(t *testing.T) {
		u := parse(t, "https://alice:hunter2@api.example.com/v1/users")
		got := redactURL(u, false)
		assert.NotContains(t, got, "hunter2")
		assert.NotContains(t, got, "alice")
		assert.Equal(t, "https://api.example.com/v1/users", got)
	})

	t.Run("sensitive query values are masked", func(t *testing.T) {
		u := parse(t, "https://api.example.com/v1?token=abc123&access_key=AKIA&page=2")
		got := redactURL(u, false)
		assert.NotContains(t, got, "abc123")
		assert.NotContains(t, got, "AKIA")
		assert.Contains(t, got, "token=%2A%2A%2A")
		assert.Contains(t, got, "access_key=%2A%2A%2A")
		assert.Contains(t, got, "page=2", "non-sensitive parameters must be preserved")
	})

	t.Run("derived sensitive keys are masked", func(t *testing.T) {
		u := parse(t, "https://api.example.com/v1?refresh_token=rt&client_secret=cs&x_api_key=k")
		got := redactURL(u, false)
		assert.NotContains(t, got, "rt")
		assert.NotContains(t, got, "cs")
		assert.NotContains(t, got, "=k")
	})

	t.Run("fragment is removed", func(t *testing.T) {
		u := parse(t, "https://api.example.com/v1/users#section")
		assert.Equal(t, "https://api.example.com/v1/users", redactURL(u, false))
	})

	// Slack Incoming Webhook のように URL のパス自体が資格情報となるケースを想定する。
	// 実サービスの Webhook URL 形式をそのままリテラルで書くと、秘密情報スキャナ
	// (GitHub Push Protection / gitleaks) が架空の値でも検知してしまうため、
	// ホストとパスは意図的に検知パターンに一致しない架空のものを使う。
	t.Run("RedactURLPath masks the credential-bearing path", func(t *testing.T) {
		u := parse(t, "https://hooks.example.com/services/WORKSPACE-ID/CHANNEL-ID/webhook-token-value")
		got := redactURL(u, true)
		assert.NotContains(t, got, "WORKSPACE-ID")
		assert.NotContains(t, got, "CHANNEL-ID")
		assert.NotContains(t, got, "webhook-token-value")
		assert.Equal(t, "https://hooks.example.com/redacted", got)
	})

	t.Run("RedactURLPath also drops the query", func(t *testing.T) {
		u := parse(t, "https://hooks.example.com/secret/path?trace=1")
		got := redactURL(u, true)
		assert.NotContains(t, got, "secret")
		assert.NotContains(t, got, "trace")
		assert.Equal(t, "https://hooks.example.com/redacted", got)
	})

	t.Run("the source URL is not mutated", func(t *testing.T) {
		raw := "https://alice:hunter2@api.example.com/v1?token=abc123"
		u := parse(t, raw)
		_ = redactURL(u, false)
		assert.Equal(t, raw, u.String(), "redactURL must operate on a copy")
	})
}
