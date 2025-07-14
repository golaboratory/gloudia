package jpcalendar

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// jikkan は十干を表すマップです。キーは0から9の整数、値は対応する漢字です。
var jikkan = map[int]string{
	0: "庚",
	1: "辛",
	2: "壬",
	3: "癸",
	4: "甲",
	5: "乙",
	6: "丙",
	7: "丁",
	8: "戊",
	9: "己",
}

// junishi は十二支を表すマップです。キーは0から11の整数、値は対応する漢字です。
var junishi = map[int]string{
	0:  "申",
	1:  "酉",
	2:  "戌",
	3:  "亥",
	4:  "子",
	5:  "丑",
	6:  "寅",
	7:  "卯",
	8:  "辰",
	9:  "巳",
	10: "午",
	11: "未",
}

// monthlyAdjustmentValues は各月ごとの調整値を格納したマップです。
// 干支日計算の補正に使用します。
var monthlyAdjustmentValues = map[int]int{
	1:  0,
	2:  31,
	3:  59,
	4:  30,
	5:  0,
	6:  31,
	7:  1,
	8:  32,
	9:  3,
	10: 33,
	11: 4,
	12: 34,
}

// GregorianYearToEtoYearString は西暦年を干支（十干十二支）の文字列に変換します。
// year: 西暦年（0以上）
// 戻り値: 干支の文字列（例: "甲子"）、またはエラー
func GregorianYearToEtoYearString(year int) (string, error) {
	if year < 0 {
		return "", fmt.Errorf("年は0以上でなければなりません")
	}
	kan := jikkan[year%10]
	shi := junishi[year%12]
	return fmt.Sprintf("%s%s", kan, shi), nil
}

// GregorianDateToEtoYearString は指定した日付の年を干支（十干十二支）の文字列に変換します。
// dt: 日付（time.Time型）
// 戻り値: 干支の文字列、またはエラー
func GregorianDateToEtoYearString(dt time.Time) (string, error) {
	year := dt.Year()
	return GregorianYearToEtoYearString(year)
}

// GregorianDateToEtoDayString は指定した日付を干支日（十干十二支）の文字列に変換します。
// dt: 日付（time.Time型）
// 戻り値: 干支日の文字列、またはエラー
func GregorianDateToEtoDayString(dt time.Time) (string, error) {
	year := dt.Year()
	month := dt.Month()
	day := dt.Day()

	foo, isLeap := calculateYear(year)
	cc := calculateCenturyConstant(year)
	y := foo + cc

	m := monthlyAdjustmentValues[int(month)]
	if isLeap && month <= 2 {
		m -= 1 // 閏年の2月までの調整
	}
	dayOfYear := y + m + day

	kan := jikkan[dayOfYear%10]
	shi := junishi[dayOfYear%12]

	return fmt.Sprintf("%s%s", kan, shi), nil
}

// calculateYear は西暦年から干支日計算用の値と閏年判定を返します。
// year: 西暦年
// 戻り値: 計算値、閏年かどうかの真偽値
func calculateYear(year int) (int, bool) {
	isLeapYear := false

	foo := year % 100
	if foo == 0 {
		if year%400 == 0 {
			return 0, true
		}
		return 0, false
	}
	bar := (foo * 10) / 2
	baz := bar / 10
	qux := bar + (baz / 2)

	isLeapYear = !strings.HasSuffix(strconv.Itoa(bar), "5") && (baz%2 == 0)

	return qux, isLeapYear
}

// calculateCenturyConstant は西暦年から世紀定数を計算して返します。
// year: 西暦年
// 戻り値: 世紀定数
func calculateCenturyConstant(year int) int {

	y := year / 100
	z := y / 4
	b := ((y * 44) + z + 13) % 60

	return b

}
