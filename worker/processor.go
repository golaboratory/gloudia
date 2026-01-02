package worker

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/newmo-oss/ergo"
)

// JobPayload はジョブの引数（JSON）をマッピングする汎用構造体です。
// 具体的なジョブの実装内で、このマップや構造体へデコードして使用します。
type JobPayload map[string]any

// Processor はジョブ実行ロジックを管理する構造体です。
// ジョブタイプとそれに対応する JobProcessor のマッピングを保持します。
type Processor struct {
	Processies map[string]JobProcessor
}

// JobProcessor は個別のジョブ処理ロジックを実装するためのインターフェースです。
// 各ジョブタイプごとにこのインターフェースを実装し、Processorに登録します。
type JobProcessor interface {
	// Process はジョブの実際の処理を行います。
	// payloadJSON にはジョブの引数がJSON形式で渡されます。
	Process(ctx context.Context, jobType string, payloadJSON json.RawMessage) error
}

// NewProcessor は新しい Processor を作成します。
// processors: ジョブタイプをキー、対応する処理実装を値とするマップ
func NewProcessor(processors map[string]JobProcessor) *Processor {
	return &Processor{Processies: processors}
}

// Process はジョブタイプに応じて処理を振り分けます。
// 登録された JobProcessor の中から jobType に一致するものを探し、実行します。
// 未知のジョブタイプの場合はエラーを返します。
func (p *Processor) Process(ctx context.Context, jobType string, payloadJSON json.RawMessage) error {
	slog.InfoContext(ctx, "Processing job", "type", jobType)

	if processor, exists := p.Processies[jobType]; exists {
		return processor.Process(ctx, jobType, payloadJSON)
	}

	return ergo.New("unknown job type", slog.String("jobType", jobType))

}
