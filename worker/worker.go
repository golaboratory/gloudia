package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// Config はワーカーの設定です。
type Config struct {
	Interval time.Duration // ポーリング間隔
}

// Worker はジョブキューシステムとのやり取りを抽象化するインターフェースです。
// 具体的な実装（PostgreSQLなど）はこのインターフェースを満たす必要があります。
type Worker interface {
	// PopNextJob は実行待ちの次のジョブを取得し、処理中ステータスに更新します。
	// ジョブが存在しない場合は sql.ErrNoRows などのエラーを返すことが期待されます。
	PopNextJob(ctx context.Context) (json.RawMessage, error)

	// ParseJob は取得したジョブのJSONデータから、ジョブIDとジョブタイプを抽出します。
	ParseJob(ctx context.Context, jobJson json.RawMessage) (int64, string, error)

	// FailJob はジョブの処理が失敗した際に呼び出され、ステータスを失敗に更新し、エラー結果を保存します。
	FailJob(ctx context.Context, jobID int64, result json.RawMessage) error

	// CompleteJob はジョブの処理が成功した際に呼び出され、ステータスを完了に更新し、結果を保存します。
	CompleteJob(ctx context.Context, jobID int64, result json.RawMessage) error
}

// WorkerProcess は非同期ジョブを実行するワーカープロセスです。
// 定期的にジョブキューをポーリングし、登録されたプロセッサーを使用してジョブを処理します。
type WorkerProcess struct {
	Worker    Worker
	processor *Processor
	cfg       Config
}

// NewWorker は新しいワーカープロセスを作成します。
// worker: ジョブキュー操作の実装
// cfg: ワーカーの設定（ポーリング間隔など）
// jobs: ジョブタイプと処理関数のマッピング
func NewWorker(worker Worker, cfg Config, jobs map[string]JobProcessor) *WorkerProcess {
	return &WorkerProcess{
		Worker:    worker,
		processor: NewProcessor(jobs),
		cfg:       cfg,
	}
}

// Start はワーカーを開始し、Contextがキャンセルされるまでブロックします。
// 指定された間隔（cfg.Interval）でジョブのポーリングを行います。
func (w *WorkerProcess) Start(ctx context.Context) {
	slog.Info("Starting background worker...")
	ticker := time.NewTicker(w.cfg.Interval)
	defer ticker.Stop()

	for {
		// まず1回処理を試みる
		processed := w.processNextJob(ctx)

		if processed {
			// 処理できた場合は、待機せずに次を見に行く
			// ただし、無限ループでCPU占有を防ぐため、コンテキストチェックを行う
			select {
			case <-ctx.Done():
				slog.Info("Stopping background worker...")
				return
			default:
				continue // 即次へ
			}
		}

		// 処理しなかった（キュー空）場合は、次のTickまで待つ
		select {
		case <-ctx.Done():
			slog.Info("Stopping background worker...")
			return
		case <-ticker.C:
			// wait
		}
	}
}

// processNextJob はDBから次のジョブを取得して実行します。
// 戻り値 bool: ジョブを処理した場合は true, ジョブがなかった場合やエラー時は false
func (w *WorkerProcess) processNextJob(ctx context.Context) bool {

	jsonJob, err := w.Worker.PopNextJob(ctx)

	if err != nil && jsonJob == nil {
		if errors.Is(err, sql.ErrNoRows) {
			// ジョブがない場合は何もしない
			return false
		}
		slog.Error("Failed to fetch job", "error", err)
		return false
	}

	// 3. 処理実行
	var resultJSON json.RawMessage
	jobID, jobType, err := w.Worker.ParseJob(ctx, jsonJob)
	if err != nil {
		slog.Error("Failed to parse job", "error", err)
		// IDが取得できている場合は失敗ステータスに更新する
		if jobID != 0 {
			errResult, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("failed to parse job: %v", err)})
			_ = w.Worker.FailJob(ctx, jobID, json.RawMessage(errResult))
		}
		return true
	}

	processErr := w.processor.Process(ctx, jobType, jsonJob)

	if processErr != nil {
		slog.ErrorContext(ctx, "Job failed", "id", jobID, "error", processErr)
		// エラー内容をResultに保存
		errResult, _ := json.Marshal(map[string]string{"error": processErr.Error()})
		resultJSON = json.RawMessage(errResult)

		updateErr := w.Worker.FailJob(ctx, jobID, resultJSON)
		if updateErr != nil {
			slog.Error("Failed to update job status", "id", jobID, "error", updateErr)
		}
	} else {
		// 成功時の結果 (必要であれば戻り値を保存)
		successResult, _ := json.Marshal(map[string]bool{"success": true})
		resultJSON = json.RawMessage(successResult)

		updateErr := w.Worker.CompleteJob(ctx, jobID, resultJSON)

		if updateErr != nil {
			slog.Error("Failed to update job status", "id", jobID, "error", updateErr)
		}
	}

	return true
}
