package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactSensitive(t *testing.T) {
	cases := []struct {
		name           string
		in             string
		mustNotContain string
		mustContain    string
	}{
		{"json password", `{"password":"secret123","name":"alice"}`, "secret123", `"password":"***"`},
		{"json access_token", `{"access_token":"xyz"}`, "xyz", `"access_token":"***"`},
		{"query token", `token=abc123&user=bob`, "abc123", "token=***"},
		{"non-sensitive preserved", `{"name":"alice"}`, "", "alice"},

		// キー名の前後に語が付く派生キーもマスクされること (完全一致だと素通りしていた)
		{"json new_password", `{"new_password":"hunter2"}`, "hunter2", `"new_password":"***"`},
		{"json current_password", `{"current_password":"hunter2"}`, "hunter2", `"current_password":"***"`},
		{"json password_confirmation", `{"password_confirmation":"hunter2"}`, "hunter2", `"password_confirmation":"***"`},
		{"json client_secret", `{"client_secret":"shhh"}`, "shhh", `"client_secret":"***"`},
		// 値に PEM 形式のヘッダ行をそのまま書くと秘密情報スキャナが
		// 架空の値でも検知しうるため、マスク対象キーの検証には無害な値を使う。
		{"json private_key", `{"private_key":"key-material-here"}`, "key-material-here", `"private_key":"***"`},
		{"json session_id", `{"session_id":"abc"}`, `"abc"`, `"session_id":"***"`},
		{"json credential", `{"aws_credential":"AKIA"}`, "AKIA", `"aws_credential":"***"`},
		{"query new_password", `new_password=hunter2&user=bob`, "hunter2", "new_password=***"},
		{"query client_secret", `client_secret=shhh`, "shhh", "client_secret=***"},

		// 値の中にキーワードが現れても、キーでなければマスクしない
		{"keyword inside a value is not masked", `{"note":"my secret plan"}`, "", "my secret plan"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := redactSensitive(c.in)
			if c.mustNotContain != "" {
				assert.NotContains(t, got, c.mustNotContain)
			}
			assert.Contains(t, got, c.mustContain)
		})
	}
}
