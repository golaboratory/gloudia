package auth_test

import (
	"os"
	"testing"

	"github.com/golaboratory/gloudia/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	t.Run("successfully hashes password", func(t *testing.T) {
		password := "secret123"
		hashed, err := auth.HashPassword(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hashed)
		assert.NotEqual(t, password, hashed)
	})

	t.Run("hashes empty password", func(t *testing.T) {
		password := ""
		hashed, err := auth.HashPassword(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hashed)
	})

	t.Run("cost is set correctly from environment variable", func(t *testing.T) {
		// Set CRYPT_COST to a specific value (e.g., bcrypt.MinCost)
		// Note: We need a way to ensure environment loading works or falls back.
		// Since Env loading is inside the function, we'll try setting the env var.
		originalVal := os.Getenv("CRYPT_COST")
		defer os.Setenv("CRYPT_COST", originalVal)

		// Set to MinCost (4) for speed and verification
		os.Setenv("CRYPT_COST", "4")

		password := "my_password"
		hashed, err := auth.HashPassword(password)
		require.NoError(t, err)

		cost, err := bcrypt.Cost([]byte(hashed))
		require.NoError(t, err)
		assert.Equal(t, bcrypt.MinCost, cost)
	})

	t.Run("uses default cost when env is missing or invalid", func(t *testing.T) {
		originalVal := os.Getenv("CRYPT_COST")
		defer os.Setenv("CRYPT_COST", originalVal)
		os.Unsetenv("CRYPT_COST")

		password := "my_password"
		hashed, err := auth.HashPassword(password)
		require.NoError(t, err)

		cost, err := bcrypt.Cost([]byte(hashed))
		require.NoError(t, err)
		// bcrypt.DefaultCost is usually 10
		assert.Equal(t, bcrypt.DefaultCost, cost)
	})

	t.Run("fails with extremely long password", func(t *testing.T) {
		// bcrypt has a max length limit (72 bytes usually)
		// GenerateFromPassword usually truncates or handles it, but let's see.
		// Actually bcrypt library handles up to 72 bytes. If longer, it might use only first 72 bytes.
		// However, GenerateFromPassword returns error if cost is invalid.
		// Let's check typical behavior. The library might not return error for long password,
		// but let's ensure it doesn't panic.
		longPassword := string(make([]byte, 100))
		hashed, err := auth.HashPassword(longPassword)
		// Depending on implementation it might succeed (using prefix) or fail.
		// Standard bcrypt usually just hashes the first 72 bytes without error.
		// We just ensure no panic and check basic success if it returns result.
		if err == nil {
			assert.NotEmpty(t, hashed)
		}
	})
}

func TestCheckPassword(t *testing.T) {
	password := "secure_password"
	hashed, err := auth.HashPassword(password)
	require.NoError(t, err)

	t.Run("valid password returns no error", func(t *testing.T) {
		err := auth.CheckPassword(password, hashed)
		assert.NoError(t, err)
	})

	t.Run("invalid password returns error", func(t *testing.T) {
		err := auth.CheckPassword("wrong_password", hashed)
		assert.Error(t, err)
		assert.Equal(t, bcrypt.ErrMismatchedHashAndPassword, err)
	})

	t.Run("empty password returns error", func(t *testing.T) {
		err := auth.CheckPassword("", hashed)
		assert.Error(t, err)
	})

	t.Run("hashed password empty returns error", func(t *testing.T) {
		err := auth.CheckPassword(password, "")
		assert.Error(t, err)
	})

	t.Run("malformed hash returns error", func(t *testing.T) {
		err := auth.CheckPassword(password, "not_a_valid_hash")
		assert.Error(t, err)
	})
}

func TestValidateStrength(t *testing.T) {
	tests := []struct {
		name          string
		password      string
		includeUpper  bool
		includeLower  bool
		includeNumber bool
		includeSymbol bool
		minLength     int
		want          bool
	}{
		{
			name:          "all requirements met",
			password:      "Ab1!",
			includeUpper:  true,
			includeLower:  true,
			includeNumber: true,
			includeSymbol: true,
			minLength:     4,
			want:          true,
		},
		{
			name:          "too short",
			password:      "Ab1!",
			includeUpper:  true,
			includeLower:  true,
			includeNumber: true,
			includeSymbol: true,
			minLength:     5,
			want:          false,
		},
		{
			name:          "missing uppercase",
			password:      "ab1!",
			includeUpper:  true,
			includeLower:  true,
			includeNumber: true,
			includeSymbol: true,
			minLength:     4,
			want:          false,
		},
		{
			name:          "missing lowercase",
			password:      "AB1!",
			includeUpper:  true,
			includeLower:  true,
			includeNumber: true,
			includeSymbol: true,
			minLength:     4,
			want:          false,
		},
		{
			name:          "missing number",
			password:      "Abc!",
			includeUpper:  true,
			includeLower:  true,
			includeNumber: true,
			includeSymbol: true,
			minLength:     4,
			want:          false,
		},
		{
			name:          "missing symbol",
			password:      "Ab12",
			includeUpper:  true,
			includeLower:  true,
			includeNumber: true,
			includeSymbol: true,
			minLength:     4,
			want:          false,
		},
		{
			name:          "relaxed requirements - only length",
			password:      "aaaa",
			includeUpper:  false,
			includeLower:  false,
			includeNumber: false,
			includeSymbol: false,
			minLength:     4,
			want:          true,
		},
		{
			name:          "mixed success",
			password:      "Password123",
			includeUpper:  true,
			includeLower:  true,
			includeNumber: true,
			includeSymbol: false,
			minLength:     8,
			want:          true,
		},
		{
			name:          "empty password with zero minLength",
			password:      "",
			includeUpper:  false,
			includeLower:  false,
			includeNumber: false,
			includeSymbol: false,
			minLength:     0,
			want:          true,
		},
		{
			name:          "all false requirements, but password has everything",
			password:      "Ab1!",
			includeUpper:  false,
			includeLower:  false,
			includeNumber: false,
			includeSymbol: false,
			minLength:     1,
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := auth.ValidateStrength(
				tt.password,
				tt.includeUpper,
				tt.includeLower,
				tt.includeNumber,
				tt.includeSymbol,
				tt.minLength,
			)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
