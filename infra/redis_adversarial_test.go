package infra

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedisClient_PoolSizeFromEnv(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	t.Run("valid env overrides default", func(t *testing.T) {
		t.Setenv("REDIS_POOL_SIZE", "25")
		client, err := NewRedisClient(mr.Addr(), "", 0)
		require.NoError(t, err)
		defer client.Close()
		assert.Equal(t, 25, client.Options().PoolSize)
	})

	t.Run("invalid env falls back to default", func(t *testing.T) {
		t.Setenv("REDIS_POOL_SIZE", "not-a-number")
		client, err := NewRedisClient(mr.Addr(), "", 0)
		require.NoError(t, err)
		defer client.Close()
		assert.Equal(t, 10, client.Options().PoolSize)
	})

	t.Run("non-positive env falls back to default", func(t *testing.T) {
		t.Setenv("REDIS_POOL_SIZE", "0")
		client, err := NewRedisClient(mr.Addr(), "", 0)
		require.NoError(t, err)
		defer client.Close()
		assert.Equal(t, 10, client.Options().PoolSize)
	})
}
