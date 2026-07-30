package jp

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGregorianDateToWareki は改元境界を含む新暦ベースの和暦変換を検証する。
func TestGregorianDateToWareki(t *testing.T) {
	tests := []struct {
		name    string
		date    time.Time
		eraName string
		year    int
	}{
		{"明治6年 (グレゴリオ暦採用日)", time.Date(1873, 1, 1, 0, 0, 0, 0, time.UTC), "明治", 6},
		{"明治45年 (最終日)", time.Date(1912, 7, 29, 0, 0, 0, 0, time.UTC), "明治", 45},
		{"大正元年 (開始日)", time.Date(1912, 7, 30, 0, 0, 0, 0, time.UTC), "大正", 1},
		{"大正15年 (最終日)", time.Date(1926, 12, 24, 0, 0, 0, 0, time.UTC), "大正", 15},
		{"昭和元年 (開始日)", time.Date(1926, 12, 25, 0, 0, 0, 0, time.UTC), "昭和", 1},
		{"昭和64年 (最終日)", time.Date(1989, 1, 7, 0, 0, 0, 0, time.UTC), "昭和", 64},
		{"平成元年 (開始日)", time.Date(1989, 1, 8, 0, 0, 0, 0, time.UTC), "平成", 1},
		{"平成31年 (最終日)", time.Date(2019, 4, 30, 0, 0, 0, 0, time.UTC), "平成", 31},
		{"令和元年 (開始日)", time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC), "令和", 1},
		{"令和8年", time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC), "令和", 8},
		{"未来の日付 (上限なし)", time.Date(2150, 1, 1, 0, 0, 0, 0, time.UTC), "令和", 132},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			era, year, err := GregorianDateToWareki(tt.date)
			require.NoError(t, err)
			assert.Equal(t, tt.eraName, era.Name)
			assert.Equal(t, tt.year, year)
		})
	}
}

// TestGregorianDateToWareki_BeforeSupportedRange は 1873 年より前の日付が
// ErrDateBeforeWareki を返すことを検証する。
func TestGregorianDateToWareki_BeforeSupportedRange(t *testing.T) {
	dates := []time.Time{
		time.Date(1872, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(1868, 10, 23, 0, 0, 0, 0, time.UTC), // 明治開始日だが改暦前のため対象外
		time.Date(1600, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	for _, d := range dates {
		_, _, err := GregorianDateToWareki(d)
		require.Error(t, err, "date=%v", d)
		assert.True(t, errors.Is(err, ErrDateBeforeWareki),
			"errors.Is(err, ErrDateBeforeWareki) で判定可能であるべき (実エラー: %v)", err)
	}
}

// TestGregorianDateToWareki_WallClock は JST の time.Time が壁時計上の日付で
// 解釈されることを検証する (令和開始日の 0 時 JST は UTC 瞬間では平成最終日)。
func TestGregorianDateToWareki_WallClock(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	era, year, err := GregorianDateToWareki(time.Date(2019, 5, 1, 0, 0, 0, 0, jst))
	require.NoError(t, err)
	assert.Equal(t, "令和", era.Name)
	assert.Equal(t, 1, year)
}

// TestGregorianDateToWarekiString は和暦文字列表記 (元年表記を含む) を検証する。
func TestGregorianDateToWarekiString(t *testing.T) {
	tests := []struct {
		name string
		date time.Time
		want string
	}{
		{"通常の年", time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC), "令和8年7月30日"},
		{"元年表記", time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC), "令和元年5月1日"},
		{"平成元年", time.Date(1989, 1, 8, 0, 0, 0, 0, time.UTC), "平成元年1月8日"},
		{"昭和64年", time.Date(1989, 1, 7, 0, 0, 0, 0, time.UTC), "昭和64年1月7日"},
		{"1桁月日", time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), "令和8年1月2日"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GregorianDateToWarekiString(tt.date)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	// 範囲外はエラー
	_, err := GregorianDateToWarekiString(time.Date(1872, 12, 31, 0, 0, 0, 0, time.UTC))
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrDateBeforeWareki))
}
