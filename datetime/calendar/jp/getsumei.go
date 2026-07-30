package jp

import (
	"time"

	"github.com/newmo-oss/ergo"
)

// getsumei は和風月名を表す配列です（インデックス 1〜12 が睦月〜師走に対応）。
var getsumei = [13]string{"", "睦月", "如月", "弥生", "卯月", "皐月", "水無月", "文月", "葉月", "長月", "神無月", "霜月", "師走"}

// ErrMonthOutOfRange は月が範囲外（1〜12以外）の場合に返されるエラーです。
var ErrMonthOutOfRange = ergo.NewSentinel("month out of range")

// ErrorMonthOutOfRange は ErrMonthOutOfRange の旧名です。
//
// Deprecated: Go の慣習 (Err プレフィックス) に合わせた ErrMonthOutOfRange を
// 使用してください。同一のセンチネル値のため errors.Is での判定結果は変わりません。
var ErrorMonthOutOfRange = ErrMonthOutOfRange

// GregorianMonthToWafuGetsumei は月の数値（1〜12）を和風月名に変換します。
// 和風月名は本来旧暦の月の異称ですが、本関数は現代の慣用に従い新暦の月に
// そのまま対応付けます。範囲外の月には ErrMonthOutOfRange を返します。
//
// 使用例:
//
//	s, err := jp.GregorianMonthToWafuGetsumei(1) // "睦月"
func GregorianMonthToWafuGetsumei(month int) (string, error) {
	if month < 1 || month > 12 {
		return "", ErrMonthOutOfRange
	}
	return getsumei[month], nil
}

// GregorianDateToWafuGetsumei は指定した日付の月（壁時計上の月）を和風月名に変換します。
func GregorianDateToWafuGetsumei(dt time.Time) (string, error) {
	return GregorianMonthToWafuGetsumei(int(dt.Month()))
}
