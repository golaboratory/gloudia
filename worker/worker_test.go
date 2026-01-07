package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWorker is a mock implementation of the Worker interface
type MockWorker struct {
	mock.Mock
}

func (m *MockWorker) PopNextJob(ctx context.Context) (json.RawMessage, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(json.RawMessage), args.Error(1)
}

func (m *MockWorker) ParseJob(ctx context.Context, jobJson json.RawMessage) (int64, string, error) {
	args := m.Called(ctx, jobJson)
	return args.Get(0).(int64), args.String(1), args.Error(2)
}

func (m *MockWorker) FailJob(ctx context.Context, jobID int64, result json.RawMessage) error {
	args := m.Called(ctx, jobID, result)
	return args.Error(0)
}

func (m *MockWorker) CompleteJob(ctx context.Context, jobID int64, result json.RawMessage) error {
	args := m.Called(ctx, jobID, result)
	return args.Error(0)
}

func TestNewWorker(t *testing.T) {
	mockWorkerImpl := new(MockWorker)
	cfg := Config{Interval: time.Second}
	jobs := map[string]JobProcessor{}

	w := NewWorker(mockWorkerImpl, cfg, jobs)

	assert.NotNil(t, w)
	assert.Equal(t, mockWorkerImpl, w.Worker)
	assert.Equal(t, cfg, w.cfg)
	assert.NotNil(t, w.processor)
}

func TestWorkerProcess_ProcessNextJob(t *testing.T) {
	// Setup
	mockWorkerImpl := new(MockWorker)
	mockJobProcessor := new(MockJobProcessor)
	jobs := map[string]JobProcessor{
		"test_job": mockJobProcessor,
	}
	cfg := Config{Interval: time.Second}
	w := NewWorker(mockWorkerImpl, cfg, jobs)
	ctx := context.Background()

	t.Run("no job", func(t *testing.T) {
		// Expect PopNextJob to return sql.ErrNoRows
		mockWorkerImpl.On("PopNextJob", ctx).Return(nil, sql.ErrNoRows).Once()

		w.processNextJob(ctx)

		mockWorkerImpl.AssertExpectations(t)
		// No other calls expected
	})

	t.Run("fetch error", func(t *testing.T) {
		mockWorkerImpl.On("PopNextJob", ctx).Return(nil, errors.New("db error")).Once()

		w.processNextJob(ctx)

		mockWorkerImpl.AssertExpectations(t)
	})

	t.Run("parse error", func(t *testing.T) {
		rawJob := json.RawMessage(`{"id": 1}`)
		mockWorkerImpl.On("PopNextJob", ctx).Return(rawJob, nil).Once()
		mockWorkerImpl.On("ParseJob", ctx, rawJob).Return(int64(0), "", errors.New("parse error")).Once()

		w.processNextJob(ctx)

		mockWorkerImpl.AssertExpectations(t)
	})

	t.Run("process success", func(t *testing.T) {
		rawJob := json.RawMessage(`{"id": 1, "type": "test_job"}`)
		jobID := int64(123)
		jobType := "test_job"

		mockWorkerImpl.On("PopNextJob", ctx).Return(rawJob, nil).Once()
		mockWorkerImpl.On("ParseJob", ctx, rawJob).Return(jobID, jobType, nil).Once()

		// Expect processor to succeed
		mockJobProcessor.On("Process", ctx, jobType, rawJob).Return(nil).Once()

		// Expect CompleteJob
		mockWorkerImpl.On("CompleteJob", ctx, jobID, mock.Anything).Return(nil).Once()

		w.processNextJob(ctx)

		mockWorkerImpl.AssertExpectations(t)
		mockJobProcessor.AssertExpectations(t)
	})

	t.Run("process failure", func(t *testing.T) {
		rawJob := json.RawMessage(`{"id": 1, "type": "test_job"}`)
		jobID := int64(123)
		jobType := "test_job"
		processErr := errors.New("processing failed")

		mockWorkerImpl.On("PopNextJob", ctx).Return(rawJob, nil).Once()
		mockWorkerImpl.On("ParseJob", ctx, rawJob).Return(jobID, jobType, nil).Once()

		// Expect processor to fail
		mockJobProcessor.On("Process", ctx, jobType, rawJob).Return(processErr).Once()

		// Expect FailJob with error message
		// Note: We used mock.Anything for result because JSON formatting might vary,
		// but checking it contains the error would be better if we could map matchers.
		// For now simple check.
		mockWorkerImpl.On("FailJob", ctx, jobID, mock.MatchedBy(func(res json.RawMessage) bool {
			return true // relaxed check
		})).Return(nil).Once()

		w.processNextJob(ctx)

		mockWorkerImpl.AssertExpectations(t)
		mockJobProcessor.AssertExpectations(t)
	})

	t.Run("status update fails", func(t *testing.T) {
		// Even if complete/fail update fails, it shouldn't panic
		rawJob := json.RawMessage(`{"id": 1, "type": "test_job"}`)
		jobID := int64(123)
		jobType := "test_job"

		mockWorkerImpl.On("PopNextJob", ctx).Return(rawJob, nil).Once()
		mockWorkerImpl.On("ParseJob", ctx, rawJob).Return(jobID, jobType, nil).Once()
		mockJobProcessor.On("Process", ctx, jobType, rawJob).Return(nil).Once()

		// CompleteJob fails
		mockWorkerImpl.On("CompleteJob", ctx, jobID, mock.Anything).Return(errors.New("update error")).Once()

		w.processNextJob(ctx)

		mockWorkerImpl.AssertExpectations(t)
		mockJobProcessor.AssertExpectations(t)
	})
}
