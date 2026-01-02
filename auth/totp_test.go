package auth_test

import (
	"strings"
	"testing"
	"time"

	"github.com/golaboratory/gloudia/auth"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetup2FA(t *testing.T) {
	t.Run("successfully generates TOTP setup info", func(t *testing.T) {
		issuer := "Manimani"
		account := "user@example.com"

		resp, err := auth.Setup2FA(issuer, account)

		require.NoError(t, err)
		require.NotNil(t, resp)

		// Secret should be generated
		assert.NotEmpty(t, resp.Secret)

		// QRCodeURI should start with otpauth://
		assert.True(t, strings.HasPrefix(resp.QRCodeURI, "otpauth://"))
		assert.Contains(t, resp.QRCodeURI, issuer)
		assert.Contains(t, resp.QRCodeURI, account)

		// QRCodeB64 should have the data URI prefix
		assert.True(t, strings.HasPrefix(resp.QRCodeB64, "data:image/png;base64,"))
	})

	t.Run("fails with invalid issuer or account", func(t *testing.T) {
		resp, err := auth.Setup2FA("", "")
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to generate TOTP key")
	})
}

func TestVerify2FA(t *testing.T) {
	issuer := "ManimaniTest"
	account := "test@example.com"

	// 1. Setup to get a real secret
	resp, err := auth.Setup2FA(issuer, account)
	require.NoError(t, err)
	secret := resp.Secret

	t.Run("successfully verifies a valid code", func(t *testing.T) {
		// Generate a valid code for the current time
		code, err := totp.GenerateCode(secret, time.Now())
		require.NoError(t, err)

		isValid := auth.Verify2FA(secret, code)
		assert.True(t, isValid)
	})

	t.Run("fails with an invalid code", func(t *testing.T) {
		isValid := auth.Verify2FA(secret, "000000")
		assert.False(t, isValid)
	})

	t.Run("fails with an empty code", func(t *testing.T) {
		isValid := auth.Verify2FA(secret, "")
		assert.False(t, isValid)
	})

	t.Run("fails with an empty secret", func(t *testing.T) {
		// Even if the code is technically "valid" for some algorithm,
		// an empty secret should fail or at least be tested.
		isValid := auth.Verify2FA("", "123456")
		assert.False(t, isValid)
	})

	t.Run("fails with a code for a different secret", func(t *testing.T) {
		// Get another secret
		resp2, _ := auth.Setup2FA(issuer, "other@example.com")
		code2, _ := totp.GenerateCode(resp2.Secret, time.Now())

		isValid := auth.Verify2FA(secret, code2)
		assert.False(t, isValid)
	})
}
