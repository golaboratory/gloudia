package datetime

import (
	"time"

	"github.com/newmo-oss/ergo"
)

// 標準的な日付フォーマット定数
const (
	// ISO8601DateFormat は ISO8601 拡張形式の日付フォーマット (例: 2026-04-25)
	ISO8601DateFormat = "2006-01-02"
	// JPDateFormat は日本語形式の日付フォーマット (例: 2026年4月25日)
	JPDateFormat = "2006年1月2日"
)

// ErrInvalidDateFormat は ParseFlexibleDate で対応外の形式が渡された場合の
// センチネルエラーです。errors.Is(err, ErrInvalidDateFormat) で判定可能です。
var ErrInvalidDateFormat = ergo.NewSentinel("invalid date format")

// ParseFlexibleDate は ISO8601 形式 ("2006-01-02") と日本語形式 ("2006年1月2日") の
// いずれかで記述された日付文字列をパースします。両形式とも失敗した場合は
// ErrInvalidDateFormat をラップしたエラーを返します。
//
// 使用例:
//
//	t, err := datetime.ParseFlexibleDate("2026-04-25")
//	t, err := datetime.ParseFlexibleDate("2026年4月25日")
//	if errors.Is(err, datetime.ErrInvalidDateFormat) { ... }
func ParseFlexibleDate(s string) (time.Time, error) {
	if t, err := time.Parse(ISO8601DateFormat, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(JPDateFormat, s); err == nil {
		return t, nil
	}
	return time.Time{}, ergo.Wrap(ErrInvalidDateFormat, "ParseFlexibleDate failed: "+s)
}
