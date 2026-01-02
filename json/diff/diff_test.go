package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeDiff(t *testing.T) {
	tests := []struct {
		name      string
		oldJSON   string
		newJSON   string
		expected  []ChangePoint
		expectErr bool
	}{
		{
			name:     "No changes",
			oldJSON:  `{"name": "test", "age": 20}`,
			newJSON:  `{"name": "test", "age": 20}`,
			expected: []ChangePoint{},
		},
		{
			name:    "Simple modification",
			oldJSON: `{"name": "old", "age": 20}`,
			newJSON: `{"name": "new", "age": 20}`,
			expected: []ChangePoint{
				{Field: "name", OldValue: "old", NewValue: "new"},
			},
		},
		{
			name:    "Addition",
			oldJSON: `{"name": "test"}`,
			newJSON: `{"name": "test", "age": 20}`,
			expected: []ChangePoint{
				{Field: "age", OldValue: nil, NewValue: float64(20)}, // JSON numbers decode to float64
			},
		},
		{
			name:    "Deletion",
			oldJSON: `{"name": "test", "age": 20}`,
			newJSON: `{"name": "test"}`,
			expected: []ChangePoint{
				{Field: "age", OldValue: float64(20), NewValue: nil},
			},
		},
		{
			name:    "Nested change",
			oldJSON: `{"user": {"name": "old"}}`,
			newJSON: `{"user": {"name": "new"}}`,
			expected: []ChangePoint{
				{Field: "user.name", OldValue: "old", NewValue: "new"},
			},
		},
		{
			name:      "Invalid JSON",
			oldJSON:   `{invalid`,
			newJSON:   `{}`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, err := ComputeDiff([]byte(tt.oldJSON), []byte(tt.newJSON))
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Order might effectively be random due to map iteration, so we check elements
			assert.Len(t, diff, len(tt.expected))
			for _, exp := range tt.expected {
				assert.Contains(t, diff, exp)
			}
		})
	}
}
