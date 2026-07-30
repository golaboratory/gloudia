package jp

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGregorianToLunar_2050IsOutOfRange(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	// 暦テーブルは太陰太陽暦 2049 年（グレゴリオ暦 2050-01-22）まで。
	// それ以降の日付は ErrOutOfSupportedRange を返すこと。
	_, err := c.GetMonth(time.Date(2050, 2, 15, 0, 0, 0, 0, time.UTC))
	require.Error(t, err, "2050-01-23 以降の日付は範囲外エラーを返すべき")
	assert.True(t, errors.Is(err, ErrOutOfSupportedRange),
		"errors.Is(err, ErrOutOfSupportedRange) で判定可能であるべき (実エラー: %v)", err)
}

func TestToDateTime_Lunar2049MappingToJan2050(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	// 旧暦 2049 年 12 月の日付はグレゴリオ暦 2050 年 1 月に対応する。逆変換 ToDateTime が
	// これを範囲外として誤って拒否しないこと。
	// 令和(eraID=5, YearOffset=2018): 旧暦年 2049 = 令和 31。
	got, err := c.ToDateTime(31, 12, 29, 5)
	require.NoError(t, err)
	assert.Equal(t, time.Date(2050, 1, 22, 0, 0, 0, 0, time.UTC), got,
		"旧暦 2049/12/29 はグレゴリオ暦 2050-01-22 に対応するはず")

	// 旧暦 2049 年 12 月は 29 日までの小の月のため、12月30日は ErrInvalidDay。
	_, err = c.ToDateTime(31, 12, 30, 5)
	require.Error(t, err, "旧暦 2049/12/30 は存在しないためエラーを返すべき")
	assert.True(t, errors.Is(err, ErrInvalidDay),
		"errors.Is(err, ErrInvalidDay) で判定可能であるべき (実エラー: %v)", err)
}
