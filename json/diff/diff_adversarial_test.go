package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeDiff_SentinelErrorIdentity(t *testing.T) {
	t.Run("malformed old json wraps ErrOldJSONUnmarshalFailed", func(t *testing.T) {
		changes, err := ComputeDiff([]byte(`{not-json`), []byte(`{}`))
		require.Error(t, err)
		assert.Nil(t, changes)
		assert.ErrorIs(t, err, ErrOldJSONUnmarshalFailed)
		assert.NotErrorIs(t, err, ErrNewJSONUnmarshalFailed)
	})

	t.Run("malformed new json wraps ErrNewJSONUnmarshalFailed", func(t *testing.T) {
		changes, err := ComputeDiff([]byte(`{}`), []byte(`{not-json`))
		require.Error(t, err)
		assert.Nil(t, changes)
		assert.ErrorIs(t, err, ErrNewJSONUnmarshalFailed)
		assert.NotErrorIs(t, err, ErrOldJSONUnmarshalFailed)
	})
}

func TestComputeDiff_Arrays(t *testing.T) {
	t.Run("element change reports whole array at array key (no index path)", func(t *testing.T) {
		changes, err := ComputeDiff([]byte(`{"tags":["a","b"]}`), []byte(`{"tags":["a","c"]}`))
		require.NoError(t, err)
		require.Len(t, changes, 1)
		assert.Equal(t, "tags", changes[0].Field)
		assert.Equal(t, []any{"a", "b"}, changes[0].OldValue)
		assert.Equal(t, []any{"a", "c"}, changes[0].NewValue)
	})

	t.Run("reordered array is reported as a change", func(t *testing.T) {
		changes, err := ComputeDiff([]byte(`{"t":[1,2]}`), []byte(`{"t":[2,1]}`))
		require.NoError(t, err)
		require.Len(t, changes, 1)
		assert.Equal(t, "t", changes[0].Field)
	})

	t.Run("identical array reports no change", func(t *testing.T) {
		changes, err := ComputeDiff([]byte(`{"t":[1,2,3]}`), []byte(`{"t":[1,2,3]}`))
		require.NoError(t, err)
		assert.Empty(t, changes)
	})

	t.Run("array of objects diffs the whole slice, not per-object keys", func(t *testing.T) {
		changes, err := ComputeDiff([]byte(`{"items":[{"id":1}]}`), []byte(`{"items":[{"id":2}]}`))
		require.NoError(t, err)
		require.Len(t, changes, 1)
		assert.Equal(t, "items", changes[0].Field)
		assert.Equal(t, []any{map[string]any{"id": float64(1)}}, changes[0].OldValue)
		assert.Equal(t, []any{map[string]any{"id": float64(2)}}, changes[0].NewValue)
	})
}

func TestComputeDiff_TypeChange(t *testing.T) {
	t.Run("scalar to object is a single change, not descended", func(t *testing.T) {
		changes, err := ComputeDiff([]byte(`{"x":"str"}`), []byte(`{"x":{"k":1}}`))
		require.NoError(t, err)
		require.Len(t, changes, 1)
		assert.Equal(t, "x", changes[0].Field)
		assert.Equal(t, "str", changes[0].OldValue)
		assert.Equal(t, map[string]any{"k": float64(1)}, changes[0].NewValue)
	})

	t.Run("object to scalar is a single change, not descended", func(t *testing.T) {
		changes, err := ComputeDiff([]byte(`{"x":{"k":1}}`), []byte(`{"x":"str"}`))
		require.NoError(t, err)
		require.Len(t, changes, 1)
		assert.Equal(t, "x", changes[0].Field)
		assert.Equal(t, map[string]any{"k": float64(1)}, changes[0].OldValue)
		assert.Equal(t, "str", changes[0].NewValue)
	})

	t.Run("scalar to array is a single change", func(t *testing.T) {
		changes, err := ComputeDiff([]byte(`{"x":1}`), []byte(`{"x":[1]}`))
		require.NoError(t, err)
		require.Len(t, changes, 1)
		assert.Equal(t, "x", changes[0].Field)
		assert.Equal(t, float64(1), changes[0].OldValue)
		assert.Equal(t, []any{float64(1)}, changes[0].NewValue)
	})
}
