package worker

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWorkerProcess_Start_ZeroIntervalDoesNotPanic(t *testing.T) {
	mockWorkerImpl := new(MockWorker)
	// 即時キャンセルでも最初の 1 回はポーリングされ得るため Maybe で許容
	mockWorkerImpl.On("PopNextJob", mock.Anything).Return(nil, sql.ErrNoRows).Maybe()

	// Interval=0 の Config。time.NewTicker(0) は panic するため、Start 内で
	// 既定値へ補正されること（panic しないこと）を検証する。
	w := NewWorker(mockWorkerImpl, Config{Interval: 0}, map[string]JobProcessor{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assert.NotPanics(t, func() {
		w.Start(ctx)
	})
}

func TestWorkerProcess_ProcessNextJob_NilDataNilError(t *testing.T) {
	// PopNextJob が (nil, nil) を返した場合、ParseJob へ進まず false を返すこと
	// （busy-loop / nil ジョブ処理を防ぐ）。
	mockWorkerImpl := new(MockWorker)
	w := NewWorker(mockWorkerImpl, Config{Interval: time.Second}, map[string]JobProcessor{})
	ctx := context.Background()

	mockWorkerImpl.On("PopNextJob", ctx).Return(nil, nil).Once()

	assert.False(t, w.processNextJob(ctx))
	mockWorkerImpl.AssertExpectations(t) // ParseJob は呼ばれない
}
