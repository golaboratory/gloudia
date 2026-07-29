package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeDiff_NullVsAbsent(t *testing.T) {
	t.Run("null to absent (key removed) is detected", func(t *testing.T) {
		changes, err := ComputeDiff([]byte(`{"a":null,"b":1}`), []byte(`{"b":1}`))
		require.NoError(t, err)
		require.Len(t, changes, 1)
		assert.Equal(t, "a", changes[0].Field)
	})

	t.Run("absent to null (key added) is detected", func(t *testing.T) {
		changes, err := ComputeDiff([]byte(`{"b":1}`), []byte(`{"a":null,"b":1}`))
		require.NoError(t, err)
		require.Len(t, changes, 1)
		assert.Equal(t, "a", changes[0].Field)
	})

	t.Run("null to null (no change) is not reported", func(t *testing.T) {
		changes, err := ComputeDiff([]byte(`{"a":null}`), []byte(`{"a":null}`))
		require.NoError(t, err)
		assert.Empty(t, changes)
	})
}
