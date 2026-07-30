package jp

import (
	"fmt"
	"time"

	"github.com/newmo-oss/ergo"
)

// センチネルエラー定義
var (
	// ErrInvalidYear は年が0未満の場合に返されるエラーです。
	ErrInvalidYear = ergo.NewSentinel("年は0以上でなければなりません")
)

// jikkan は十干を表す配列です。インデックスは西暦年を 10 で割った剰余に対応します
// （例: 下 1 桁が 4 の年は「甲」）。
var jikkan = [10]string{"庚", "辛", "壬", "癸", "甲", "乙", "丙", "丁", "戊", "己"}

// junishi は十二支を表す配列です。インデックスは西暦年を 12 で割った剰余に対応します
// （例: 12 で割り切れる年は「申」）。
var junishi = [12]string{"申", "酉", "戌", "亥", "子", "丑", "寅", "卯", "辰", "巳", "午", "未"}

// monthlyAdjustmentValues は干支日計算に使う月ごとの調整値です（インデックス 1〜12）。
// 値は平年の各月 1 日までの通算日数を 60 で割った剰余です。
var monthlyAdjustmentValues = [13]int{0, 0, 31, 59, 30, 0, 31, 1, 32, 3, 33, 4, 34}

// GregorianYearToEtoYearString は西暦年を年干支（十干十二支）の文字列に変換します。
// 年の区切りはグレゴリオ暦の 1 月 1 日です（立春や旧正月を区切りとする流儀とは
// 異なります）。year が負の場合は ErrInvalidYear をラップしたエラーを返します。
//
// 使用例:
//
//	s, err := jp.GregorianYearToEtoYearString(2024) // "甲辰"
func GregorianYearToEtoYearString(year int) (string, error) {
	if year < 0 {
		return "", ergo.Wrap(ErrInvalidYear, fmt.Sprintf("%d", year))
	}
	return jikkan[year%10] + junishi[year%12], nil
}

// GregorianDateToEtoYearString は指定した日付の年を干支（十干十二支）の文字列に変換します。
// dt: 日付（time.Time型）
// 戻り値: 干支の文字列、またはエラー
func GregorianDateToEtoYearString(dt time.Time) (string, error) {
	year := dt.Year()
	return GregorianYearToEtoYearString(year)
}

// GregorianDateToEtoDayString は指定した日付（グレゴリオ暦）を
// 日干支（十干十二支）の文字列に変換します。日付は dt の壁時計上の年月日で
// 解釈します。年が負の場合は ErrInvalidYear をラップしたエラーを返します。
//
// 使用例:
//
//	s, err := jp.GregorianDateToEtoDayString(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) // "甲子"
func GregorianDateToEtoDayString(dt time.Time) (string, error) {
	year := dt.Year()
	// 負の年は計算過程で負の剰余となり、十干十二支の参照が範囲外となるため拒否する。
	// GregorianYearToEtoYearString と同じ契約に揃える。
	if year < 0 {
		return "", ergo.Wrap(ErrInvalidYear, fmt.Sprintf("%d", year))
	}
	month := dt.Month()
	day := dt.Day()

	yearVal, isLeap := calculateYear(year)
	cc := calculateCenturyConstant(year)
	y := yearVal + cc

	m := monthlyAdjustmentValues[int(month)]
	if isLeap && month <= 2 {
		m -= 1 // 年定数は 3 月以降に合わせてあるため、閏年の 1〜2 月は 1 日戻す
	}
	dayOfYear := y + m + day

	return jikkan[dayOfYear%10] + junishi[dayOfYear%12], nil
}

// calculateYear は西暦年の下 2 桁から日干支計算用の年内定数と、
// その年がグレゴリオ暦の閏年かどうかを返します。
// 年内定数は y を下 2 桁として 5y + y/4 (整数除算) です。
func calculateYear(year int) (int, bool) {
	yearMod100 := year % 100
	if yearMod100 == 0 {
		return 0, year%400 == 0
	}
	return 5*yearMod100 + yearMod100/4, yearMod100%4 == 0
}

// calculateCenturyConstant は西暦年の世紀部分から日干支計算用の定数を返します。
func calculateCenturyConstant(year int) int {
	c := year / 100
	return ((c * 44) + c/4 + 13) % 60
}
