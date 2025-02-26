package log_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	slog "log/slog"

	"github.com/golaboratory/gloudia/core/log"
)

// TestNewLogger_Text は、Textハンドラーを用いて Logger が正しく生成され、ログ出力できることを検証します。
func TestNewLogger_Text(t *testing.T) {
	// os.Stdout を一時的にパイプに置き換えて出力をキャプチャ
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("パイプ作成エラー: %v", err)
	}
	os.Stdout = w

	logger := log.New(log.Level(slog.LevelInfo), log.Text)
	if logger == nil {
		t.Fatal("Logger が nil です")
	}

	// ログ出力
	logger.Info("test message", slog.String("key", "value"))

	// パイプを閉じ、出力内容を取得
	w.Close()
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("出力取得エラー: %v", err)
	}
	os.Stdout = origStdout

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("出力に 'test message' が含まれていません: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("出力に 'key=value' が含まれていません: %s", output)
	}
}

// TestNewLogger_Json は、Jsonハンドラーを用いて Logger が正しく生成され、ログ出力できることを検証します。
func TestNewLogger_Json(t *testing.T) {
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("パイプ作成エラー: %v", err)
	}
	os.Stdout = w

	logger := log.New(log.Level(slog.LevelInfo), log.Json)
	if logger == nil {
		t.Fatal("Logger が nil です")
	}

	// ログ出力
	logger.Info("json test", slog.String("foo", "bar"))

	// パイプを閉じ、出力内容を取得
	w.Close()
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("出力取得エラー: %v", err)
	}
	os.Stdout = origStdout

	output := buf.String()
	if !strings.Contains(output, "json test") {
		t.Errorf("出力に 'json test' が含まれていません: %s", output)
	}
	if !strings.Contains(output, `"foo":"bar"`) {
		t.Errorf("出力に '\"foo\":\"bar\"' が含まれていません: %s", output)
	}
}
