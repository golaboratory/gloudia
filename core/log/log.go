package log

import (
	"log/slog"
	"os"
)

// Level はloggerの出力レベルを表す型です。
type Level slog.Level

// Handler はログのフォーマットタイプを表す列挙型です。
//   - Text: テキスト形式
//   - Json: JSON形式
type Handler int

const (
	// Text はテキスト形式のログ出力を指定します。
	Text Handler = iota
	// Json はJSON形式のログ出力を指定します。
	Json
)

// Logger はslog.Loggerをラップしたカスタムロガーです。
type Logger struct {
	*slog.Logger
}

// New は指定されたレベルとハンドラータイプに基づいて新しいLoggerを生成します。
// 引数:
//   - level: ログ出力の最小レベル
//   - handler: ログ出力形式（TextまたはJson）
//
// 戻り値:
//   - *Logger: 生成されたLoggerインスタンス
func New(level Level, handler Handler) *Logger {
	opts := &slog.HandlerOptions{
		Level: slog.Level(level),
	}

	var internalHandler slog.Handler

	switch handler {
	case Text:
		internalHandler = slog.NewTextHandler(os.Stdout, opts)
	case Json:
		internalHandler = slog.NewJSONHandler(os.Stdout, opts)
	}

	l := Logger{
		Logger: slog.New(internalHandler),
	}

	return &l
}
