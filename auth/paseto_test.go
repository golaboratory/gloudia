package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasetoMaker(t *testing.T) {
	maker, err := NewTokenMaker(GenerateRandomKey())
	require.NoError(t, err)

	userID := int64(123)
	tenantID := "test-tenant"
	roleID := int64(1)
	duration := time.Minute

	issuedAt := time.Now()
	_ = issuedAt.Add(duration)

	token, err := maker.CreateToken(userID, tenantID, roleID, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	assert.Equal(t, userID, payload.UserID)
	assert.Equal(t, tenantID, payload.TenantID)
	assert.Equal(t, roleID, payload.RoleID)

	// Test expired token
	// This is a bit hard to test deterministically without sleep or mocking time,
	// but standard library time.Sleep is acceptable for unit test of expiration

	// Create token with very short duration
	tokenShort, err := maker.CreateToken(userID, tenantID, roleID, -time.Minute) // Already expired
	require.NoError(t, err)

	payloadExpired, err := maker.VerifyToken(tokenShort)
	assert.Error(t, err)
	assert.Nil(t, payloadExpired)
}

func TestNewTokenMaker_InvalidKey(t *testing.T) {
	_, err := NewTokenMaker("invalid-key-size")
	assert.Error(t, err)
}
