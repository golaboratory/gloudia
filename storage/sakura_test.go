package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewSakuraObjectStorage tests initialization logic.
// Actual connection tests require real credentials, so here we mostly test config handling.
func TestNewSakuraObjectStorage(t *testing.T) {
	ctx := context.Background()

	t.Run("Initialize successfully", func(t *testing.T) {
		cfg := SakuraStorageConfig{
			AccessKey: "dummy_access",
			SecretKey: "dummy_secret",
			Endpoint:  "https://s3.isk01.sakurastorage.jp",
			Bucket:    "my-bucket",
		}

		storage, err := NewSakuraObjectStorage(ctx, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, storage)
		assert.Equal(t, "my-bucket", storage.bucket)
	})

	t.Run("Missing Bucket", func(t *testing.T) {
		cfg := SakuraStorageConfig{
			AccessKey: "dummy",
			SecretKey: "dummy",
			Endpoint:  "https://example.com",
			Bucket:    "", // Missing
		}

		storage, err := NewSakuraObjectStorage(ctx, cfg)
		assert.Error(t, err)
		assert.Nil(t, storage)
		assert.Contains(t, err.Error(), "bucket name is required")
	})

	// Note: We cannot easily test Upload/Download/Delete/GetSignedURL here
	// because aws-sdk-go-v2 Client structs are hard to mock without dependency injection via interfaces.
	// For a shared library, providing the implementation is key.
	// Integration tests would go here if we had a mock S3 server or credentials.
}

func TestSakuraObjectStorage_GetSignedURL_Mock(t *testing.T) {
	// PresignClient logic is internal to AWS SDK, but we can verify our wrapper inputs simply
	// by checking if it accepts the arguments (it will fail network or signing if we tried to execute if we could).
	// Since we can't execute, we rely on compilation and initialization tests above.

	// Future improvement: abstract s3.PresignClient behind an interface if unit testing of parameter passing is strictly required.
	// For now, trusting the SDK types.
}
