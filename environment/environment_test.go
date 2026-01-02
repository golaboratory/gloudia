package environment_test

import (
	"testing"
	"time"

	"github.com/golaboratory/gloudia/environment"
)

type Config struct {
	StringVal   string        `envconfig:"TEST_STRING"`
	IntVal      int           `envconfig:"TEST_INT"`
	BoolVal     bool          `envconfig:"TEST_BOOL"`
	DurationVal time.Duration `envconfig:"TEST_DURATION"`
	SliceVal    []string      `envconfig:"TEST_SLICE"`
}

type DefaultConfig struct {
	Val string `envconfig:"TEST_DEFAULT" default:"default_value"`
}

type RequiredConfig struct {
	Val string `envconfig:"TEST_REQUIRED" required:"true"`
}

type NestedRoot struct {
	Server ServerConfig
}

type ServerConfig struct {
	Port int `envconfig:"TEST_SERVER_PORT"`
}

func TestNewEnvValue(t *testing.T) {
	t.Run("BasicTypes", func(t *testing.T) {
		t.Setenv("TEST_STRING", "hello")
		t.Setenv("TEST_INT", "123")
		t.Setenv("TEST_BOOL", "true")
		t.Setenv("TEST_DURATION", "10s")
		t.Setenv("TEST_SLICE", "a,b,c")

		cfg, err := environment.NewEnvValue[Config]()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.StringVal != "hello" {
			t.Errorf("expected StringVal to be 'hello', got '%s'", cfg.StringVal)
		}
		if cfg.IntVal != 123 {
			t.Errorf("expected IntVal to be 123, got %d", cfg.IntVal)
		}
		if cfg.BoolVal != true {
			t.Errorf("expected BoolVal to be true, got %v", cfg.BoolVal)
		}
		if cfg.DurationVal != 10*time.Second {
			t.Errorf("expected DurationVal to be 10s, got %v", cfg.DurationVal)
		}
		if len(cfg.SliceVal) != 3 || cfg.SliceVal[0] != "a" || cfg.SliceVal[1] != "b" || cfg.SliceVal[2] != "c" {
			t.Errorf("expected SliceVal to be [a b c], got %v", cfg.SliceVal)
		}
	})

	t.Run("DefaultValues", func(t *testing.T) {
		// 環境変数を設定しない
		cfg, err := environment.NewEnvValue[DefaultConfig]()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Val != "default_value" {
			t.Errorf("expected Val to be 'default_value', got '%s'", cfg.Val)
		}
	})

	t.Run("RequiredError", func(t *testing.T) {
		// 必須環境変数を設定しない
		_, err := environment.NewEnvValue[RequiredConfig]()
		if err == nil {
			t.Fatal("expected error for missing required variable, got nil")
		}
	})

	t.Run("RequiredSuccess", func(t *testing.T) {
		t.Setenv("TEST_REQUIRED", "present")
		cfg, err := environment.NewEnvValue[RequiredConfig]()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Val != "present" {
			t.Errorf("expected Val to be 'present', got '%s'", cfg.Val)
		}
	})

	t.Run("TypeConversionError", func(t *testing.T) {
		t.Setenv("TEST_INT", "not_an_int")
		_, err := environment.NewEnvValue[Config]()
		if err == nil {
			t.Fatal("expected error for invalid type conversion, got nil")
		}
	})

	t.Run("NestedStructure", func(t *testing.T) {
		// envconfigはネストされた構造体をサポートしていますが、
		// Process("", &v) の場合、プレフィックスなしでフィールドのタグを直接見に行きます。
		// ただし、ネストされたフィールドにタグがない場合、親フィールド名_子フィールド名になる挙動もありますが、
		// 今回は明示的にタグをつけているケースを確認します。
		t.Setenv("TEST_SERVER_PORT", "8080")

		cfg, err := environment.NewEnvValue[NestedRoot]()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Server.Port != 8080 {
			t.Errorf("expected Server.Port to be 8080, got %d", cfg.Server.Port)
		}
	})

	t.Run("NonStructType", func(t *testing.T) {
		// envconfigは構造体のポインタを期待するため、プリミティブ型を渡すとエラーになることが予想されます
		_, err := environment.NewEnvValue[int]()
		if err == nil {
			t.Fatal("expected error when T is not a struct, got nil")
		}
	})
}
