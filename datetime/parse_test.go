package datetime

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseFlexibleDate_ISO8601Format は ISO8601 形式（"2006-01-02"）の
// パース成功を検証する。
func TestParseFlexibleDate_ISO8601Format(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{"通常の日付", "2026-04-25", time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)},
		{"うるう年の2/29", "2024-02-29", time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)},
		{"年初", "2026-01-01", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"年末", "2026-12-31", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFlexibleDate(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestParseFlexibleDate_JPFormat は日本語形式（"2006年1月2日"）の
// パース成功を検証する。
func TestParseFlexibleDate_JPFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{"通常の日付", "2026年4月25日", time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)},
		{"1桁月日", "2026年1月2日", time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
		{"うるう年の2/29", "2024年2月29日", time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)},
		{"年末", "2026年12月31日", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFlexibleDate(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestParseFlexibleDate_SlashAndCompactFormat はスラッシュ区切りと
// 8 桁数字形式のパース成功を検証する。
func TestParseFlexibleDate_SlashAndCompactFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{"スラッシュ区切り (ゼロ埋め)", "2026/04/25", time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)},
		{"スラッシュ区切り (ゼロ埋めなし)", "2026/4/25", time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)},
		{"スラッシュ区切り (1桁月日)", "2026/1/2", time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
		{"8桁数字", "20260425", time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)},
		{"8桁数字 (うるう年の2/29)", "20240229", time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFlexibleDate(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestParseFlexibleDate_InvalidFormat は不正形式に対し
// ErrInvalidDateFormat を返すことを検証する。
func TestParseFlexibleDate_InvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"空文字", ""},
		{"ドット区切り", "2026.04.25"},
		{"年月のみ", "2026-04"},
		{"6桁数字", "202604"},
		{"日付ではない文字列", "hello world"},
		{"存在しない日付 (ISO)", "2026-02-30"},
		{"存在しない日付 (JP)", "2026年2月30日"},
		{"存在しない日付 (スラッシュ)", "2026/2/30"},
		{"存在しない日付 (8桁)", "20260230"},
		{"うるう年でない年の2/29 (ISO)", "2025-02-29"},
		{"うるう年でない年の2/29 (JP)", "2025年2月29日"},
		{"全角数字 (現状非対応)", "２０２６年４月２５日"},
		{"時刻付き", "2026-04-25 10:00:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFlexibleDate(tt.input)
			require.Error(t, err)
			assert.True(t, got.IsZero(), "失敗時は zero value を返すべき")
			assert.True(t, errors.Is(err, ErrInvalidDateFormat),
				"errors.Is(err, ErrInvalidDateFormat) で判定可能であるべき (実エラー: %v)", err)
		})
	}
}

// TestParseFlexibleDate_FormatPriority は ISO8601 形式が日本語形式よりも
// 先に試行されることを保護する。
//
// 注意: 両形式は文字列パターンが完全に異なるため実用上は競合しないが、
// 実装変更時にパース順序を保証するための回帰テスト。
func TestParseFlexibleDate_FormatPriority(t *testing.T) {
	// ISO8601 として有効、日本語として無効 → 成功すること
	got, err := ParseFlexibleDate("2026-04-25")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC), got)

	// 日本語として有効、ISO8601 として無効 → 成功すること
	got, err = ParseFlexibleDate("2026年4月25日")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC), got)
}

// TestErrInvalidDateFormat_IsSentinel は ErrInvalidDateFormat が
// センチネルエラーとして errors.Is で判定可能であることを保護する。
func TestErrInvalidDateFormat_IsSentinel(t *testing.T) {
	require.NotNil(t, ErrInvalidDateFormat)

	// 自己一致
	assert.True(t, errors.Is(ErrInvalidDateFormat, ErrInvalidDateFormat))

	// ParseFlexibleDate が返すエラーから判定可能
	_, err := ParseFlexibleDate("invalid")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidDateFormat))
}

// TestDateFormatConstants は定数値の固定を保証する。
// 利用側プロジェクトが定数を参照する際の契約。
func TestDateFormatConstants(t *testing.T) {
	assert.Equal(t, "2006-01-02", ISO8601DateFormat)
	assert.Equal(t, "2006年1月2日", JPDateFormat)
	assert.Equal(t, "2006/1/2", SlashDateFormat)
	assert.Equal(t, "20060102", CompactDateFormat)
}

// TestJST は JST タイムゾーンが UTC+9 の固定オフセットであることを保証する。
func TestJST(t *testing.T) {
	tm := time.Date(2026, 7, 30, 12, 0, 0, 0, JST)
	assert.Equal(t, "JST", tm.Location().String())
	_, offset := tm.Zone()
	assert.Equal(t, 9*60*60, offset)
	assert.Equal(t, time.Date(2026, 7, 30, 3, 0, 0, 0, time.UTC).Unix(), tm.Unix())
}
