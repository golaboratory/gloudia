package log

import (
	"log/slog"
	"os"
)

// Level は logger の出力レベルを表します。
type Level slog.Level

// Handler はログのフォーマットタイプを表す列挙型です。
type Handler int

const (
	// Text はテキスト形式のログ出力を指定します。
	Text Handler = iota
	// Json はJSON形式のログ出力を指定します。
	Json
)

// Logger は slog.Logger をラップしたカスタムロガーです。
type Logger struct {
	*slog.Logger
}

// New は指定されたレベルとハンドラータイプに基づいて新しい Logger を生成します。
// 引数:
//   - level: ログ出力の最小レベル。
//   - handler: ログ出力形式（Text または Json）。
//
// 戻り値:
//   - *Logger: 生成された Logger インスタンス。
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
