package excel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertColumnIndexToLetter(t *testing.T) {
	// 0始まりインデックス → Excel 列名（bijective base-26）。
	// 旧実装は 702("AAA") 以降で誤った列名を返していた。
	cases := map[int]string{
		0:     "A",
		25:    "Z",
		26:    "AA",
		701:   "ZZ",
		702:   "AAA",
		728:   "ABA",
		16383: "XFD", // 最終列
	}
	for idx, want := range cases {
		got, err := convertColumnIndexToLetter(idx)
		require.NoError(t, err, "index=%d", idx)
		assert.Equal(t, want, got, "index=%d", idx)
	}

	// 範囲外はエラー
	_, err := convertColumnIndexToLetter(-1)
	assert.ErrorIs(t, err, ColumnIndexOutOfRangeError)
	_, err = convertColumnIndexToLetter(16384) // MaxColumns（0始まりでは範囲外）
	assert.ErrorIs(t, err, ColumnIndexOutOfRangeError)
}
