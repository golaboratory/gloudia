package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockWorker / MockJobProcessor are defined in worker_test.go and processor_test.go
// (same package); they are reused here and intentionally not redefined.

func TestWorkerProcess_ProcessNextJob_ParseErrorWithJobIDMarksFailed(t *testing.T) {
	mockWorkerImpl := new(MockWorker)
	w := NewWorker(mockWorkerImpl, Config{Interval: time.Second}, map[string]JobProcessor{})
	ctx := context.Background()

	rawJob := json.RawMessage(`{"id": 7}`)
	parseErr := errors.New("bad payload shape")

	mockWorkerImpl.On("PopNextJob", ctx).Return(rawJob, nil).Once()
	mockWorkerImpl.On("ParseJob", ctx, rawJob).Return(int64(7), "", parseErr).Once()

	var captured json.RawMessage
	mockWorkerImpl.On("FailJob", ctx, int64(7), mock.MatchedBy(func(res json.RawMessage) bool {
		captured = res
		return true
	})).Return(nil).Once()

	// ParseJob 失敗でも jobID が取れていれば true を返し、FailJob で失敗マークされる。
	processed := w.processNextJob(ctx)
	assert.True(t, processed)

	var payload map[string]string
	require.NoError(t, json.Unmarshal(captured, &payload))
	assert.Contains(t, payload["error"], "failed to parse job")
	assert.Contains(t, payload["error"], "bad payload shape")

	mockWorkerImpl.AssertExpectations(t)
}

func TestWorkerProcess_ProcessNextJob_FailResultContainsProcessError(t *testing.T) {
	mockWorkerImpl := new(MockWorker)
	mockJobProcessor := new(MockJobProcessor)
	jobs := map[string]JobProcessor{"test_job": mockJobProcessor}
	w := NewWorker(mockWorkerImpl, Config{Interval: time.Second}, jobs)
	ctx := context.Background()

	rawJob := json.RawMessage(`{"id": 1, "type": "test_job"}`)
	processErr := errors.New("processing boom")

	mockWorkerImpl.On("PopNextJob", ctx).Return(rawJob, nil).Once()
	mockWorkerImpl.On("ParseJob", ctx, rawJob).Return(int64(123), "test_job", nil).Once()
	mockJobProcessor.On("Process", ctx, "test_job", rawJob).Return(processErr).Once()

	var captured json.RawMessage
	mockWorkerImpl.On("FailJob", ctx, int64(123), mock.MatchedBy(func(res json.RawMessage) bool {
		captured = res
		return true
	})).Return(nil).Once()

	processed := w.processNextJob(ctx)
	assert.True(t, processed)

	var payload map[string]string
	require.NoError(t, json.Unmarshal(captured, &payload))
	assert.Equal(t, "processing boom", payload["error"])

	mockWorkerImpl.AssertExpectations(t)
	mockJobProcessor.AssertExpectations(t)
}

func TestWorkerProcess_ProcessNextJob_CompleteResultIsSuccessTrue(t *testing.T) {
	mockWorkerImpl := new(MockWorker)
	mockJobProcessor := new(MockJobProcessor)
	jobs := map[string]JobProcessor{"test_job": mockJobProcessor}
	w := NewWorker(mockWorkerImpl, Config{Interval: time.Second}, jobs)
	ctx := context.Background()

	rawJob := json.RawMessage(`{"id": 1, "type": "test_job"}`)

	mockWorkerImpl.On("PopNextJob", ctx).Return(rawJob, nil).Once()
	mockWorkerImpl.On("ParseJob", ctx, rawJob).Return(int64(123), "test_job", nil).Once()
	mockJobProcessor.On("Process", ctx, "test_job", rawJob).Return(nil).Once()

	var captured json.RawMessage
	mockWorkerImpl.On("CompleteJob", ctx, int64(123), mock.MatchedBy(func(res json.RawMessage) bool {
		captured = res
		return true
	})).Return(nil).Once()

	processed := w.processNextJob(ctx)
	assert.True(t, processed)

	var payload map[string]bool
	require.NoError(t, json.Unmarshal(captured, &payload))
	assert.True(t, payload["success"])

	mockWorkerImpl.AssertExpectations(t)
	mockJobProcessor.AssertExpectations(t)
}

func TestWorkerProcess_ProcessNextJob_FetchErrorReturnsFalse(t *testing.T) {
	mockWorkerImpl := new(MockWorker)
	w := NewWorker(mockWorkerImpl, Config{Interval: time.Second}, map[string]JobProcessor{})
	ctx := context.Background()

	// sql.ErrNoRows 以外のフェッチエラーは false を返し、次の Tick まで待機する（busy-loop 防止）。
	mockWorkerImpl.On("PopNextJob", ctx).Return(nil, errors.New("db down")).Once()

	assert.False(t, w.processNextJob(ctx))
	mockWorkerImpl.AssertExpectations(t)
}
