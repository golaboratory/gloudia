package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateStrength_CountsRunesNotBytes(t *testing.T) {
	// 3文字の日本語パスワード（UTF-8では9バイト）。minLength=5 で拒否されること。
	ok, err := ValidateStrength("あいう", false, false, false, false, 5)
	assert.NoError(t, err)
	assert.False(t, ok, "3 文字は minLength=5 を満たさない（バイト数で誤判定しないこと）")

	// 5文字なら許可されること。
	ok2, err := ValidateStrength("あいうえお", false, false, false, false, 5)
	assert.NoError(t, err)
	assert.True(t, ok2)
}
