package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"
)

// Config はワーカーの設定です。
type Config struct {
	Interval time.Duration // ポーリング間隔
}

// DefaultConfig はデフォルトのワーカー設定を返します。
// ポーリング間隔は5秒に設定されています。
func DefaultConfig() Config {
	return Config{
		Interval: 5 * time.Second,
	}
}

// Worker はジョブキューシステムとのやり取りを抽象化するインターフェースです。
// 具体的な実装（PostgreSQLなど）はこのインターフェースを満たす必要があります。
type Worker interface {
	// PopNextJob は実行待ちの次のジョブを取得し、処理中ステータスに更新します。
	// ジョブが存在しない場合は sql.ErrNoRows などのエラーを返すことが期待されます。
	PopNextJob(ctx context.Context) (json.RawMessage, error)

	// ParseJob は取得したジョブのJSONデータから、ジョブIDとジョブタイプを抽出します。
	// パースに失敗した場合でも、可能な限り jobID を返すこと。jobID が返れば
	// ワーカーが当該ジョブを失敗として記録でき、'processing' 状態での滞留を防げる。
	// jobID を抽出できないパース失敗のために、実装側で古い 'processing' ジョブを
	// 回収する reaper を用意することを推奨する。
	ParseJob(ctx context.Context, jobJSON json.RawMessage) (int64, string, error)

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
	// 不正な Interval（0 以下）は time.NewTicker を panic させるため、既定値に補正する。
	if w.cfg.Interval <= 0 {
		slog.Warn("invalid worker interval; falling back to default", slog.Duration("default", DefaultConfig().Interval))
		w.cfg.Interval = DefaultConfig().Interval
	}
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
//
// ユーザー提供の ParseJob / JobProcessor.Process が panic しても、ワーカーの
// ポーリングループ (Start) が停止しないよう、ここで panic を回復する。
// panic 発生時はスタックトレースを記録し false を返す (次の Tick まで待機して
// 再開することで、回復不能なジョブによるビジーループを避ける)。
func (w *WorkerProcess) processNextJob(ctx context.Context) (processed bool) {
	var jobID int64
	defer func() {
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "panic recovered while processing job",
				slog.Any("recover", r),
				slog.String("stack", string(debug.Stack())))
			// パニックしたジョブを 'processing' のまま放置しないよう、ジョブIDが
			// 判明していれば best-effort で失敗としてマークする (FailJob 自体の
			// 失敗/パニックはワーカーを巻き込まないよう内側の recover で無視する)。
			if jobID != 0 {
				func() {
					defer func() { _ = recover() }()
					errResult, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("job processor panicked: %v", r)})
					if failErr := w.Worker.FailJob(ctx, jobID, json.RawMessage(errResult)); failErr != nil {
						slog.ErrorContext(ctx, "Failed to mark panicked job as failed", "id", jobID, "error", failErr)
					}
				}()
			}
			processed = false
		}
	}()

	jsonJob, err := w.Worker.PopNextJob(ctx)

	// フェッチエラーは（部分データの有無に関わらず）常にハンドリングする。
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// ジョブがない場合は何もしない（次の Tick まで待機）
			return false
		}
		slog.ErrorContext(ctx, "Failed to fetch job", "error", err)
		return false
	}
	// エラーもデータも無い場合はキュー空とみなし、次の Tick まで待機する（busy-loop 防止）。
	if jsonJob == nil {
		return false
	}

	// 3. 処理実行
	var resultJSON json.RawMessage
	jobID, jobType, err := w.Worker.ParseJob(ctx, jsonJob)
	if err != nil {
		// IDが取得できている場合は失敗ステータスに更新する
		if jobID != 0 {
			slog.ErrorContext(ctx, "Failed to parse job", "id", jobID, "error", err)
			errResult, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("failed to parse job: %v", err)})
			if failErr := w.Worker.FailJob(ctx, jobID, json.RawMessage(errResult)); failErr != nil {
				slog.ErrorContext(ctx, "Failed to update job status after parse error", "id", jobID, "error", failErr)
			}
		} else {
			// jobID が取得できないため FailJob で失敗マークできず、ジョブは
			// PopNextJob により 'processing' へ遷移したまま滞留する恐れがある。
			// 実装側は ParseJob が失敗時も可能な限り jobID を返すか、一定時間
			// 'processing' のままのジョブを回収する reaper を用意すること。
			slog.ErrorContext(ctx, "Failed to parse job and no job ID could be extracted; the job may be stranded in 'processing' state — implement a reaper for stale jobs or have ParseJob surface the job ID", "error", err)
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
