package service

import (
	"testing"

	"github.com/golaboratory/gloudia/api/config"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	bs := BaseService{}
	// LoadConfig内でエラーがあれば標準出力に出力されるのみなので、panicが発生しなければ成功とする
	bs.LoadConfig()
	// APIConfigが初期化されていることを簡易的に検証（詳細はconfig.ApiConfigの実装に依存）
	assert.IsType(t, config.ApiConfig{}, bs.APIConfig)
}

func TestAddInvalid(t *testing.T) {
	bs := BaseService{}
	bs.AddInvalid("id", "Id is required")
	assert.Equal(t, 1, len(bs.InvalidList))
	assert.Equal(t, "id", bs.InvalidList[0].Name)
	assert.Equal(t, "Id is required", bs.InvalidList[0].Message)
}

func TestIsValid(t *testing.T) {
	bs := BaseService{}
	// 初期状態ではエラーがないためIsValidはtrue
	assert.True(t, bs.IsValid())
	// エラーを追加するとfalseになる
	bs.AddInvalid("id", "Id is required")
	assert.False(t, bs.IsValid())
}

func TestClearInvalid(t *testing.T) {
	bs := BaseService{}
	bs.AddInvalid("id", "Id is required")
	bs.AddInvalid("name", "Name is required")
	assert.False(t, bs.IsValid())
	// ClearInvalidでエラーリストをクリア
	bs.ClearInvalid()
	assert.True(t, bs.IsValid())
	assert.Equal(t, 0, len(bs.InvalidList))
}
