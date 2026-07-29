package jp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGregorianToLunar_2101IsOutOfRange(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	// 暦テーブルは太陰太陽暦 2100 年までのため、グレゴリオ暦 2101 年の日付は変換できず
	// エラーを返すこと（旧実装は不正な月や負日を nil エラーで返していた）。
	_, err := c.GetMonth(time.Date(2101, 1, 15, 0, 0, 0, 0, time.UTC))
	assert.Error(t, err, "2101 の日付は範囲外エラーを返すべき")
}

func TestToDateTime_Lunar2100MappingToJan2101(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	// 旧暦 2100 年 12 月の日付はグレゴリオ暦 2101 年 1 月に対応する。逆変換 ToDateTime が
	// これを範囲外として誤って拒否しないこと（maxSupportedDate 引き下げによる regression 防止）。
	// 令和(eraID=5, YearOffset=2018): 旧暦年 2100 = 令和 82。
	got, err := c.ToDateTime(82, 12, 5, 5)
	require.NoError(t, err)
	assert.Equal(t, 2101, got.Year(), "旧暦 2100/12 はグレゴリオ暦 2101 年に対応するはず")
}
