package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockJobProcessor is a mock implementation of JobProcessor
type MockJobProcessor struct {
	mock.Mock
}

func (m *MockJobProcessor) Process(ctx context.Context, jobType string, payloadJSON json.RawMessage) error {
	args := m.Called(ctx, jobType, payloadJSON)
	return args.Error(0)
}

func TestNewProcessor(t *testing.T) {
	mockProcessor := new(MockJobProcessor)
	jobs := map[string]JobProcessor{
		"test_job": mockProcessor,
	}

	processor := NewProcessor(jobs)
	assert.NotNil(t, processor)
	assert.Equal(t, jobs, processor.Processors)
}

func TestProcessor_Process(t *testing.T) {
	mockProcessor := new(MockJobProcessor)
	jobs := map[string]JobProcessor{
		"test_job": mockProcessor,
	}
	processor := NewProcessor(jobs)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		payload := json.RawMessage(`{"key":"value"}`)
		mockProcessor.On("Process", ctx, "test_job", payload).Return(nil).Once()

		err := processor.Process(ctx, "test_job", payload)
		assert.NoError(t, err)
		mockProcessor.AssertExpectations(t)
	})

	t.Run("unknown job type", func(t *testing.T) {
		payload := json.RawMessage(`{}`)
		err := processor.Process(ctx, "unknown_job", payload)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown job type")
	})

	t.Run("processor error", func(t *testing.T) {
		payload := json.RawMessage(`{}`)
		expectedErr := errors.New("processing failed")
		mockProcessor.On("Process", ctx, "test_job", payload).Return(expectedErr).Once()

		err := processor.Process(ctx, "test_job", payload)
		assert.ErrorIs(t, err, expectedErr)
		mockProcessor.AssertExpectations(t)
	})
}
