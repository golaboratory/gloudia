package jp

import (
	"errors"
	"time"
)

// getsumei は和風月名を表すマップです。
// キーは1〜12の整数、値は対応する和風月名（睦月〜師走）です。
var getsumei = map[int]string{
	1:  "睦月",
	2:  "如月",
	3:  "弥生",
	4:  "卯月",
	5:  "皐月",
	6:  "水無月",
	7:  "文月",
	8:  "葉月",
	9:  "長月",
	10: "神無月",
	11: "霜月",
	12: "師走",
}

// ErrorMonthOutOfRange は月が範囲外（1〜12以外）の場合に返されるエラーです。
var (
	ErrorMonthOutOfRange = errors.New("month out of range")
)

// GregorianMonthToWafuGetsumei は西暦の月（1〜12）を和風月名に変換します。
// month: 月（1〜12）
// 戻り値: 和風月名（例: "睦月"）、または範囲外の場合はエラー
func GregorianMonthToWafuGetsumei(month int) (string, error) {
	if month < 1 || month > 12 {
		return "", ErrorMonthOutOfRange
	}
	return getsumei[month], nil
}

// GregorianDateToWafuGetsumei は指定した日付の月を和風月名に変換します。
// dt: 日付（time.Time型）
// 戻り値: 和風月名、または範囲外の場合はエラー
func GregorianDateToWafuGetsumei(dt time.Time) (string, error) {
	return GregorianMonthToWafuGetsumei(int(dt.Month()))
}
