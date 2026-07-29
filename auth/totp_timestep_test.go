package auth

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
)

func TestVerify2FAWithTimeStep(t *testing.T) {
	setup, err := Setup2FA("Issuer", "user@example.com")
	assert.NoError(t, err)

	code, err := totp.GenerateCode(setup.Secret, time.Now())
	assert.NoError(t, err)

	ok, step, err := Verify2FAWithTimeStep(setup.Secret, code)
	assert.NoError(t, err)
	assert.True(t, ok, "現在のコードは検証に成功すること")
	assert.Greater(t, step, int64(0), "一致したタイムステップが返ること（リプレイ判定用）")

	// 明らかに不正なコードは失敗すること。
	bad, _, err := Verify2FAWithTimeStep(setup.Secret, "abcdef")
	assert.NoError(t, err)
	assert.False(t, bad)
}
