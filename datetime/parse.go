package datetime

import (
	"time"

	"github.com/newmo-oss/ergo"
)

// JST は日本標準時 (UTC+9) の固定タイムゾーンです。
// 日本に夏時間は存在しないため、システムの tzdata (Asia/Tokyo) に依存しない
// time.FixedZone による定義で正確です。
var JST = time.FixedZone("JST", 9*60*60)

// 標準的な日付フォーマット定数
const (
	// ISO8601DateFormat は ISO8601 拡張形式の日付フォーマット (例: 2026-04-25)
	ISO8601DateFormat = "2006-01-02"
	// JPDateFormat は日本語形式の日付フォーマット (例: 2026年4月25日)
	JPDateFormat = "2006年1月2日"
	// SlashDateFormat はスラッシュ区切りの日付フォーマット (例: 2026/4/25, 2026/04/25)
	SlashDateFormat = "2006/1/2"
	// CompactDateFormat は 8 桁数字の日付フォーマット (例: 20260425)
	CompactDateFormat = "20060102"
)

// flexibleDateFormats は ParseFlexibleDate が試行するフォーマットの優先順リストです。
var flexibleDateFormats = []string{
	ISO8601DateFormat,
	JPDateFormat,
	SlashDateFormat,
	CompactDateFormat,
}

// ErrInvalidDateFormat は ParseFlexibleDate で対応外の形式が渡された場合の
// センチネルエラーです。errors.Is(err, ErrInvalidDateFormat) で判定可能です。
var ErrInvalidDateFormat = ergo.NewSentinel("invalid date format")

// ParseFlexibleDate は ISO8601 形式 ("2006-01-02")、日本語形式 ("2006年4月2日")、
// スラッシュ区切り ("2006/4/2", "2006/04/02")、8 桁数字 ("20060402") のいずれかで
// 記述された日付文字列をパースします。日本語形式とスラッシュ区切りの月日は
// ゼロ埋めなしでも受け付けます。戻り値は UTC の 0 時の time.Time です。
// すべての形式で失敗した場合は ErrInvalidDateFormat をラップしたエラーを返します。
//
// 使用例:
//
//	t, err := datetime.ParseFlexibleDate("2026-04-25")
//	t, err := datetime.ParseFlexibleDate("2026年4月25日")
//	t, err := datetime.ParseFlexibleDate("2026/04/25")
//	t, err := datetime.ParseFlexibleDate("20260425")
//	if errors.Is(err, datetime.ErrInvalidDateFormat) { ... }
func ParseFlexibleDate(s string) (time.Time, error) {
	for _, format := range flexibleDateFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, ergo.Wrap(ErrInvalidDateFormat, "ParseFlexibleDate failed: "+s)
}
