package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCryptor_DecryptTamperedCiphertextSameKey(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	c, err := NewCryptor(key)
	require.NoError(t, err)

	enc, err := c.Encrypt("integrity-protected payload")
	require.NoError(t, err)

	raw, err := base64.URLEncoding.DecodeString(enc)
	require.NoError(t, err)
	// raw = nonce(12) + ciphertext + tag(16); must be well past the nonce.
	require.Greater(t, len(raw), 12)

	t.Run("flip last byte (tag)", func(t *testing.T) {
		tampered := make([]byte, len(raw))
		copy(tampered, raw)
		tampered[len(tampered)-1] ^= 0xFF
		_, err := c.Decrypt(base64.URLEncoding.EncodeToString(tampered))
		require.Error(t, err)
		// Auth failure must NOT masquerade as a length error.
		assert.False(t, errors.Is(err, ErrCiphertextTooShort))
	})

	t.Run("flip a ciphertext byte", func(t *testing.T) {
		tampered := make([]byte, len(raw))
		copy(tampered, raw)
		// Index 12 is the first ciphertext byte (right after the 12-byte nonce).
		tampered[12] ^= 0x01
		_, err := c.Decrypt(base64.URLEncoding.EncodeToString(tampered))
		require.Error(t, err)
	})

	t.Run("flip a nonce byte", func(t *testing.T) {
		tampered := make([]byte, len(raw))
		copy(tampered, raw)
		tampered[0] ^= 0x01
		_, err := c.Decrypt(base64.URLEncoding.EncodeToString(tampered))
		require.Error(t, err)
	})
}

func TestCryptor_DecryptNonceSizeBoundary(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	c, err := NewCryptor(key)
	require.NoError(t, err)

	tests := []struct {
		name         string
		len          int
		wantTooShort bool
	}{
		{"empty", 0, true},
		{"nonceSize minus one", 11, true},
		{"exactly nonceSize", 12, false},  // passes length check, fails GCM open
		{"nonceSize plus one", 13, false}, // passes length check, fails GCM open
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := base64.URLEncoding.EncodeToString(make([]byte, tt.len))
			_, err := c.Decrypt(in)
			require.Error(t, err)
			if tt.wantTooShort {
				assert.ErrorIs(t, err, ErrCiphertextTooShort)
			} else {
				// Past the length guard: must be a GCM auth/decrypt failure, not the length sentinel.
				assert.False(t, errors.Is(err, ErrCiphertextTooShort))
			}
		})
	}
}

func TestNewCryptor_InvalidKeyLengthsSentinel(t *testing.T) {
	for _, n := range []int{0, 1, 15, 16, 24, 31, 33, 64} {
		t.Run("len_"+strconv.Itoa(n), func(t *testing.T) {
			c, err := NewCryptor(make([]byte, n))
			assert.Nil(t, c)
			require.Error(t, err)
			// Must surface the exported sentinel, not just any error string.
			assert.ErrorIs(t, err, ErrInvalidKeyLength)
		})
	}

	t.Run("nil key", func(t *testing.T) {
		c, err := NewCryptor(nil)
		assert.Nil(t, c)
		assert.ErrorIs(t, err, ErrInvalidKeyLength)
	})

	t.Run("exactly 32 bytes succeeds", func(t *testing.T) {
		c, err := NewCryptor(make([]byte, 32))
		require.NoError(t, err)
		assert.NotNil(t, c)
	})
}

func TestCryptor_EmptyPlaintextAndNonceUniqueness(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	c, err := NewCryptor(key)
	require.NoError(t, err)

	t.Run("empty plaintext round-trips", func(t *testing.T) {
		enc, err := c.Encrypt("")
		require.NoError(t, err)
		assert.NotEmpty(t, enc) // nonce + tag are still emitted
		dec, err := c.Decrypt(enc)
		require.NoError(t, err)
		assert.Equal(t, "", dec)
	})

	t.Run("nonce is unique per call for identical plaintext", func(t *testing.T) {
		const msg = "repeat me"
		seen := make(map[string]struct{}, 50)
		for i := 0; i < 50; i++ {
			enc, err := c.Encrypt(msg)
			require.NoError(t, err)
			if _, dup := seen[enc]; dup {
				t.Fatalf("duplicate ciphertext for identical plaintext on iteration %d: nonce reuse", i)
			}
			seen[enc] = struct{}{}
			// Each still decrypts back to the same plaintext.
			dec, err := c.Decrypt(enc)
			require.NoError(t, err)
			assert.Equal(t, msg, dec)
		}
	})
}

func TestCryptor_ConcurrentEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	c, err := NewCryptor(key)
	require.NoError(t, err)

	const goroutines = 32
	const iters = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)
	errCh := make(chan error, goroutines*iters)
	for g := 0; g < goroutines; g++ {
		go func(g int) {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				msg := fmt.Sprintf("g=%d i=%d 秘密", g, i)
				enc, e := c.Encrypt(msg)
				if e != nil {
					errCh <- e
					return
				}
				dec, e := c.Decrypt(enc)
				if e != nil {
					errCh <- e
					return
				}
				if dec != msg {
					errCh <- fmt.Errorf("round-trip mismatch: got %q want %q", dec, msg)
					return
				}
			}
		}(g)
	}
	wg.Wait()
	close(errCh)
	for e := range errCh {
		t.Fatalf("concurrent failure: %v", e)
	}
}
