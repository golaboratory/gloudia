package auth

import (
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// --- paseto adversarial tests ---

func TestVerifyToken_WrongKeyRejected(t *testing.T) {
	makerA, err := NewTokenMaker(GenerateRandomKey())
	require.NoError(t, err)
	makerB, err := NewTokenMaker(GenerateRandomKey())
	require.NoError(t, err)

	token, err := makerA.CreateToken(int64(7), "tenant-x", int64(3), time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// A token issued by makerA must NOT verify under makerB's key.
	payload, err := makerB.VerifyToken(token)
	require.Error(t, err)
	assert.Nil(t, payload)
	assert.ErrorIs(t, err, ErrTokenVerificationFailed)

	// Sanity: the same token still verifies under the correct key.
	good, err := makerA.VerifyToken(token)
	require.NoError(t, err)
	require.NotNil(t, good)
	assert.Equal(t, int64(7), good.UserID)
}

func TestVerifyToken_TamperedTokenRejected(t *testing.T) {
	maker, err := NewTokenMaker(GenerateRandomKey())
	require.NoError(t, err)

	token, err := maker.CreateToken(int64(42), "tenant-z", int64(9), time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Flip one character in the payload portion of the token to corrupt the
	// authenticated ciphertext. v4.local format is: v4.local.<payload>(.footer)
	b := []byte(token)
	// Mutate a byte well past the "v4.local." prefix so we hit the ciphertext.
	idx := len(b) - 5
	require.Greater(t, idx, len("v4.local."))
	if b[idx] == 'A' {
		b[idx] = 'B'
	} else {
		b[idx] = 'A'
	}
	tampered := string(b)
	require.NotEqual(t, token, tampered)

	payload, err := maker.VerifyToken(tampered)
	require.Error(t, err)
	assert.Nil(t, payload)
	assert.ErrorIs(t, err, ErrTokenVerificationFailed)
}

func TestNewTokenMaker_ErrorSentinels(t *testing.T) {
	tests := []struct {
		name    string
		hexKey  string
		wantErr error
	}{
		{
			name:    "empty key",
			hexKey:  "",
			wantErr: ErrInvalidKeySize,
		},
		{
			name:    "too short",
			hexKey:  "abcd",
			wantErr: ErrInvalidKeySize,
		},
		{
			name:    "too long (66 chars)",
			hexKey:  strings.Repeat("a", 66),
			wantErr: ErrInvalidKeySize,
		},
		{
			// Correct length (64) but contains non-hex characters -> decode fails.
			name:    "correct length but invalid hex",
			hexKey:  strings.Repeat("z", 64),
			wantErr: ErrInvalidHexKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maker, err := NewTokenMaker(tt.hexKey)
			require.Error(t, err)
			assert.Nil(t, maker)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}

	// Positive control: a freshly generated key is accepted.
	maker, err := NewTokenMaker(GenerateRandomKey())
	require.NoError(t, err)
	assert.NotNil(t, maker)
}

func TestVerifyTokenWithExpiry_ReturnsExpiration(t *testing.T) {
	maker, err := NewTokenMaker(GenerateRandomKey())
	require.NoError(t, err)

	duration := 5 * time.Minute
	before := time.Now()
	token, err := maker.CreateToken(int64(11), "tenant-exp", int64(2), duration)
	require.NoError(t, err)
	after := time.Now()

	payload, exp, err := maker.VerifyTokenWithExpiry(token)
	require.NoError(t, err)
	require.NotNil(t, payload)
	assert.Equal(t, int64(11), payload.UserID)
	assert.Equal(t, "tenant-exp", payload.TenantID)
	assert.Equal(t, int64(2), payload.RoleID)

	// exp must be (issue time + duration), within the window the token was created.
	assert.False(t, exp.IsZero(), "expiration must be populated")
	assert.WithinRange(t, exp, before.Add(duration).Add(-2*time.Second), after.Add(duration).Add(2*time.Second))
}

func TestVerifyToken_InvalidClaimTypes(t *testing.T) {
	maker, err := NewTokenMaker(GenerateRandomKey())
	require.NoError(t, err)

	build := func(userID, tenantID, roleID any) string {
		tok := paseto.NewToken()
		tok.SetIssuedAt(time.Now())
		tok.SetNotBefore(time.Now())
		tok.SetExpiration(time.Now().Add(time.Minute))
		tok.Set("user_id", userID)
		tok.Set("tenant_id", tenantID)
		tok.Set("role_id", roleID)
		return tok.V4Encrypt(maker.symmetricKey, nil)
	}

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{
			name:    "user_id wrong type",
			token:   build("not-a-number", "tenant", int64(1)),
			wantErr: ErrInvalidTokenPayloadUserID,
		},
		{
			name:    "tenant_id wrong type",
			token:   build(int64(1), int64(999), int64(1)),
			wantErr: ErrInvalidTokenPayloadTenantID,
		},
		{
			name:    "role_id wrong type",
			token:   build(int64(1), "tenant", "not-a-number"),
			wantErr: ErrInvalidTokenPayloadRoleID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := maker.VerifyToken(tt.token)
			require.Error(t, err)
			assert.Nil(t, payload)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestGenerateRandomKey_Contract(t *testing.T) {
	k1 := GenerateRandomKey()
	k2 := GenerateRandomKey()

	assert.Len(t, k1, 64, "key must be 64 hex characters (32 bytes)")
	assert.NotEqual(t, k1, k2, "successive keys must differ")

	// Must be valid hex.
	_, err := hex.DecodeString(k1)
	require.NoError(t, err)

	// Must be directly usable by NewTokenMaker.
	maker, err := NewTokenMaker(k1)
	require.NoError(t, err)
	assert.NotNil(t, maker)
}

// --- password adversarial tests ---

func TestGenerateRandomCode_Behavior(t *testing.T) {
	t.Run("length zero returns ErrInvalidLength", func(t *testing.T) {
		code, err := GenerateRandomCode(0, true, true, true, true)
		require.Error(t, err)
		assert.Empty(t, code)
		assert.ErrorIs(t, err, ErrInvalidLength)
	})

	t.Run("negative length returns ErrInvalidLength", func(t *testing.T) {
		code, err := GenerateRandomCode(-3, true, false, false, false)
		require.Error(t, err)
		assert.Empty(t, code)
		assert.ErrorIs(t, err, ErrInvalidLength)
	})

	t.Run("no character types selected returns ErrNoCharacterTypesSelected", func(t *testing.T) {
		code, err := GenerateRandomCode(8, false, false, false, false)
		require.Error(t, err)
		assert.Empty(t, code)
		assert.ErrorIs(t, err, ErrNoCharacterTypesSelected)
	})

	t.Run("length shorter than required types returns ErrLengthTooShort", func(t *testing.T) {
		// 4 required types but length 3 -> too short.
		code, err := GenerateRandomCode(3, true, true, true, true)
		require.Error(t, err)
		assert.Empty(t, code)
		assert.ErrorIs(t, err, ErrLengthTooShort)
	})

	t.Run("produces requested length and includes each required type", func(t *testing.T) {
		// Run several times since selection is random; every output must satisfy the contract.
		for i := 0; i < 50; i++ {
			code, err := GenerateRandomCode(16, true, true, true, true)
			require.NoError(t, err)
			assert.Equal(t, 16, len([]rune(code)), "output rune length must equal requested length")
			// ValidateStrength must confirm all four requested types are present.
			ok, vErr := ValidateStrength(code, true, true, true, true, 16)
			require.NoError(t, vErr)
			assert.True(t, ok, "generated code %q must contain all requested character types", code)
		}
	})

	t.Run("single type only contains that type", func(t *testing.T) {
		code, err := GenerateRandomCode(10, false, false, true, false)
		require.NoError(t, err)
		assert.Equal(t, 10, len([]rune(code)))
		for _, r := range code {
			assert.Truef(t, r >= '0' && r <= '9', "expected only digits, got %q", string(r))
		}
	})

	t.Run("length equal to required count succeeds", func(t *testing.T) {
		// boundary: length == number of required types (4).
		code, err := GenerateRandomCode(4, true, true, true, true)
		require.NoError(t, err)
		assert.Equal(t, 4, len([]rune(code)))
		ok, vErr := ValidateStrength(code, true, true, true, true, 4)
		require.NoError(t, vErr)
		assert.True(t, ok)
	})
}

func TestHashPassword_72ByteTruncation(t *testing.T) {
	base := strings.Repeat("a", 72)

	hashed, err := HashPassword(base)
	require.NoError(t, err)
	require.NotEmpty(t, hashed)

	t.Run("first 72 bytes match -> differing suffix still verifies", func(t *testing.T) {
		// Same 72-byte prefix, extra bytes beyond 72 are ignored by bcrypt.
		samePrefix := base + "EXTRA-IGNORED-SUFFIX"
		err := CheckPassword(samePrefix, hashed)
		assert.NoError(t, err, "bcrypt truncates to 72 bytes; suffix must be ignored")
	})

	t.Run("differing byte within first 72 is rejected", func(t *testing.T) {
		diffWithin72 := strings.Repeat("a", 71) + "b"
		err := CheckPassword(diffWithin72, hashed)
		assert.Error(t, err)
		assert.ErrorIs(t, err, bcrypt.ErrMismatchedHashAndPassword)
	})
}

// --- totp adversarial tests ---

func TestVerify2FAWithTimeStep_SkewWindowAndDistinctSteps(t *testing.T) {
	setup, err := Setup2FA("Issuer", "skew@example.com")
	require.NoError(t, err)
	secret := setup.Secret

	opts := totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	}
	now := time.Now()

	codeFor := func(offsetSteps int) string {
		t2 := now.Add(time.Duration(offsetSteps) * 30 * time.Second)
		c, gErr := totp.GenerateCodeCustom(secret, t2, opts)
		require.NoError(t, gErr)
		return c
	}

	prevCode := codeFor(-1)
	curCode := codeFor(0)
	nextCode := codeFor(1)

	okPrev, stepPrev, err := Verify2FAWithTimeStep(secret, prevCode)
	require.NoError(t, err)
	assert.True(t, okPrev, "previous-step code within skew window must be accepted")

	okCur, stepCur, err := Verify2FAWithTimeStep(secret, curCode)
	require.NoError(t, err)
	assert.True(t, okCur, "current-step code must be accepted")

	okNext, stepNext, err := Verify2FAWithTimeStep(secret, nextCode)
	require.NoError(t, err)
	assert.True(t, okNext, "next-step code within skew window must be accepted")

	// The whole point of replay protection: distinct windows yield distinct steps.
	// These relationships are deterministic regardless of any clock drift between
	// the test and the function, because the returned step is tied to the time
	// bucket the supplied code was generated for (always three consecutive buckets).
	assert.Equal(t, stepCur-1, stepPrev, "previous window must be exactly one step earlier")
	assert.Equal(t, stepCur+1, stepNext, "next window must be exactly one step later")

	// A code two steps away from `now` is outside the +/-1 skew window relative to
	// `now`. We validate the rejection against a FIXED reference time (now) using
	// the same opts, rather than Verify2FAWithTimeStep, to avoid a flake: that
	// function samples its own time.Now(), so a code two steps ahead of the test's
	// `now` can land inside the function's window if the call crosses a period
	// boundary. ValidateCustom with an explicit time is deterministic.
	outOfWindow, gErr := totp.GenerateCodeCustom(secret, now.Add(2*30*time.Second), opts)
	require.NoError(t, gErr)
	if outOfWindow != prevCode && outOfWindow != curCode && outOfWindow != nextCode {
		valid, vErr := totp.ValidateCustom(outOfWindow, secret, now, opts)
		require.NoError(t, vErr)
		assert.False(t, valid, "code two steps from now must be outside the +/-1 skew window")
	}
}
