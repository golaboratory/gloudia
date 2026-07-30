// このファイルの太陰太陽暦（旧暦）変換ロジックおよび yearInfo テーブルは、
// .NET (dotnet/runtime) の System.Globalization.JapaneseLunisolarCalendar /
// EastAsianLunisolarCalendar の実装（内部テーブル s_yinfo）を参考に、
// Go へ移植したものです。yearInfo の 1960〜2049 年の全行が .NET の s_yinfo と
// 一致することを機械的に照合済みです。s_yinfo の暦データ自体の出典は
// Reingold & Dershowitz "Calendrical Calculations" および
// 西沢優荘『暦日大鑑』（新人物往来社, 1994）です。
//
// 参考元:
//
//	https://github.com/dotnet/runtime
//	Copyright (c) .NET Foundation and Contributors
//	Licensed under the MIT License.
//
// 元号データ（EraInfo）についても同実装の定義に合わせています。
package jp

import (
	"fmt"
	"sort"
	"time"

	"github.com/newmo-oss/ergo"
)

// センチネルエラー定義
var (
	// ErrOutOfSupportedRange は日付がサポート範囲外の場合のエラーです。
	ErrOutOfSupportedRange = ergo.NewSentinel("out of the supported range")
	// ErrOutOfLunisolarRange は年が太陰太陽暦のサポート範囲外の場合のエラーです。
	ErrOutOfLunisolarRange = ergo.NewSentinel("out of the supported lunisolar range")
	// ErrEraNotFound は日付に対応する元号が見つからない場合のエラーです。
	ErrEraNotFound = ergo.NewSentinel("could not determine era for the date")
	// ErrInternalEraNotFound は内部処理で元号が見つからない場合のエラーです。
	ErrInternalEraNotFound = ergo.NewSentinel("internal error: era not found")
	// ErrInvalidMonth は指定された月が年の月数範囲外の場合のエラーです。
	ErrInvalidMonth = ergo.NewSentinel("invalid month for year")
	// ErrInvalidDay は指定された日が月の日数範囲外の場合のエラーです。
	ErrInvalidDay = ergo.NewSentinel("invalid day for month")
	// ErrInvalidEra は無効な元号IDが指定された場合のエラーです。
	ErrInvalidEra = ergo.NewSentinel("invalid era")
)

// --- 定数とグローバル変数 ---

// サポートする和暦（太陰太陽暦）の年範囲。
// 上限の 2049 は出典である .NET の s_yinfo テーブルの範囲に一致します。
// （かつて存在した 2050〜2100 年の拡張データは出典がなく、閏月の欠落や
// 旧正月日付の連鎖矛盾を含む誤データであることが確認されたため削除しました。）
const (
	minLunisolarYear = 1960
	maxLunisolarYear = 2049
)

// サポートするグレゴリオ暦の日付範囲。
// 太陰太陽暦 1960 年の元日（1960-01-28）から 2049 年の大晦日（2050-01-22）まで。
var (
	minSupportedDate = time.Date(1960, 1, 28, 0, 0, 0, 0, time.UTC)
	maxSupportedDate = time.Date(2050, 1, 22, 23, 59, 59, 999999999, time.UTC)
)

// グレゴリオ暦の閏年判定
func isGregorianLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// daysInGregorianYear はグレゴリオ暦の年における日数を返します
func daysInGregorianYear(year int) int {
	if isGregorianLeapYear(year) {
		return 366
	}
	return 365
}

// yearInfo は .NET の EastAsianLunisolarCalendar 内部テーブル s_yinfo に対応します
// （出典はファイル冒頭の注記を参照）。
// 各行のデータ: [閏月(なければ0), 正月月, 正日, 各月の日数パターン(ビットマスク)]
// 日数パターンは最上位ビットから第1月・第2月…の順に、1 なら 30 日・0 なら 29 日を表します。
var yearInfo = [][4]int{
	{6, 1, 28, 0b1010110101010000},  // 1960
	{0, 2, 15, 0b1010101101010000},  // 1961
	{0, 2, 5, 0b0100101101100000},   // 1962
	{4, 1, 25, 0b1010010101110000},  // 1963
	{0, 2, 13, 0b1010010101110000},  // 1964
	{0, 2, 2, 0b0101001001110000},   // 1965
	{3, 1, 22, 0b0110100100110000},  // 1966
	{0, 2, 9, 0b1101100101010000},   // 1967
	{7, 1, 30, 0b0110101010101000},  // 1968
	{0, 2, 17, 0b0101011010100000},  // 1969
	{0, 2, 6, 0b1001101011010000},   // 1970
	{5, 1, 27, 0b0100101011101000},  // 1971
	{0, 2, 15, 0b0100101011100000},  // 1972
	{0, 2, 3, 0b1010010011100000},   // 1973
	{4, 1, 23, 0b1101001001101000},  // 1974
	{0, 2, 11, 0b1101001001010000},  // 1975
	{8, 1, 31, 0b1101010101001000},  // 1976
	{0, 2, 18, 0b1011010101000000},  // 1977
	{0, 2, 7, 0b1101011010100000},   // 1978
	{6, 1, 28, 0b1001011011010000},  // 1979
	{0, 2, 16, 0b1001010110110000},  // 1980
	{0, 2, 5, 0b0100100110110000},   // 1981
	{4, 1, 25, 0b1010010011011000},  // 1982
	{0, 2, 13, 0b1010010010110000},  // 1983
	{10, 2, 2, 0b1011001001011000},  // 1984
	{0, 2, 20, 0b0110101001010000},  // 1985
	{0, 2, 9, 0b0110110101000000},   // 1986
	{6, 1, 29, 0b1011010110101000},  // 1987
	{0, 2, 18, 0b0010101101100000},  // 1988
	{0, 2, 6, 0b1001010110110000},   // 1989
	{5, 1, 27, 0b0100100110111000},  // 1990
	{0, 2, 15, 0b0100100101110000},  // 1991
	{0, 2, 4, 0b0110010010110000},   // 1992
	{3, 1, 23, 0b0110101001010000},  // 1993
	{0, 2, 10, 0b1110101001010000},  // 1994
	{8, 1, 31, 0b0110110101001000},  // 1995
	{0, 2, 19, 0b0101101011010000},  // 1996
	{0, 2, 8, 0b0010101101100000},   // 1997
	{5, 1, 28, 0b1001001101110000},  // 1998
	{0, 2, 16, 0b1001001011100000},  // 1999
	{0, 2, 5, 0b1100100101100000},   // 2000
	{4, 1, 24, 0b1110010010101000},  // 2001
	{0, 2, 12, 0b1101010010100000},  // 2002
	{0, 2, 1, 0b1101101001010000},   // 2003
	{2, 1, 22, 0b0101101010101000},  // 2004
	{0, 2, 9, 0b0101011011000000},   // 2005
	{7, 1, 29, 0b1010101011011000},  // 2006
	{0, 2, 18, 0b0010010111010000},  // 2007
	{0, 2, 7, 0b1001001011010000},   // 2008
	{5, 1, 26, 0b1100100101011000},  // 2009
	{0, 2, 14, 0b1010100101010000},  // 2010
	{0, 2, 3, 0b1011010010100000},   // 2011
	{3, 1, 23, 0b1011101001010000},  // 2012
	{0, 2, 10, 0b1011010101010000},  // 2013
	{9, 1, 31, 0b0101010110101000},  // 2014
	{0, 2, 19, 0b0100101110100000},  // 2015
	{0, 2, 8, 0b1010010110110000},   // 2016
	{5, 1, 28, 0b0101001010111000},  // 2017
	{0, 2, 16, 0b0101001010110000},  // 2018
	{0, 2, 5, 0b1010100101010000},   // 2019
	{4, 1, 25, 0b1011010010101000},  // 2020
	{0, 2, 12, 0b0110101010100000},  // 2021
	{0, 2, 1, 0b1010110101010000},   // 2022
	{2, 1, 22, 0b0101010110101000},  // 2023
	{0, 2, 10, 0b0100101101100000},  // 2024
	{6, 1, 29, 0b1010010101110000},  // 2025
	{0, 2, 17, 0b1010010101110000},  // 2026
	{0, 2, 7, 0b0101001001110000},   // 2027
	{5, 1, 27, 0b0110100100110000},  // 2028
	{0, 2, 13, 0b1101100100110000},  // 2029
	{0, 2, 3, 0b0101101010100000},   // 2030
	{3, 1, 23, 0b1010101101010000},  // 2031
	{0, 2, 11, 0b1001011011010000},  // 2032
	{11, 1, 31, 0b0100101011101000}, // 2033
	{0, 2, 19, 0b0100101011100000},  // 2034
	{0, 2, 8, 0b1010010011010000},   // 2035
	{6, 1, 28, 0b1101001001101000},  // 2036
	{0, 2, 15, 0b1101001001010000},  // 2037
	{0, 2, 4, 0b1101010100100000},   // 2038
	{5, 1, 24, 0b1101101010100000},  // 2039
	{0, 2, 12, 0b1011011010100000},  // 2040
	{0, 2, 1, 0b1001011011010000},   // 2041
	{2, 1, 22, 0b0100101011011000},  // 2042
	{0, 2, 10, 0b0100100110110000},  // 2043
	{7, 1, 30, 0b1010010010111000},  // 2044
	{0, 2, 17, 0b1010010010110000},  // 2045
	{0, 2, 6, 0b1011001001010000},   // 2046
	{5, 1, 26, 0b1011010100101000},  // 2047
	{0, 2, 14, 0b0110110101000000},  // 2048
	{0, 2, 2, 0b1010110110100000},   // 2049
}

// --- 構造体定義 ---

// EraInfo は元号の情報を保持します。
type EraInfo struct {
	Era         int       // 元号ID (昭和=3, 平成=4, 令和=5。.NET の JapaneseCalendar と同じ番号)
	Name        string    // 元号名 (例: "令和")
	EnglishName string    // 元号の英語名 (例: "Reiwa")
	StartDate   time.Time // 元号の開始日 (UTC 0時)
	YearOffset  int       // グレゴリオ暦年から元号年を引いた値 (例: 令和 2019 - 1 = 2018)
}

// JapaneseLunisolarCalendar は和暦（太陰太陽暦）の機能を提供します
type JapaneseLunisolarCalendar struct {
	eras []EraInfo
}

// --- コンストラクタ ---

// NewJapaneseLunisolarCalendar はカレンダーの新しいインスタンスを作成します
func NewJapaneseLunisolarCalendar() *JapaneseLunisolarCalendar {
	// .NETの実装に合わせて元号データを定義（新しい順）
	// サポート範囲(1960-)に関連する元号に限定
	eras := []EraInfo{
		{5, "令和", "Reiwa", time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC), 2018},
		{4, "平成", "Heisei", time.Date(1989, 1, 8, 0, 0, 0, 0, time.UTC), 1988},
		{3, "昭和", "Showa", time.Date(1926, 12, 25, 0, 0, 0, 0, time.UTC), 1925},
	}
	return &JapaneseLunisolarCalendar{eras: eras}
}

// --- 内部ヘルパー関数 ---

// toUTCDate は t の壁時計上の日付（年月日）を UTC の 0 時として取り出します。
// 本パッケージの暦計算は時刻やタイムゾーンではなく「日付」に対して定義されるため、
// JST など任意のロケーションの time.Time をそのまま渡しても、その場所での
// 日付として解釈されます。
func toUTCDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func (c *JapaneseLunisolarCalendar) checkDateRange(t time.Time) error {
	d := toUTCDate(t)
	if d.Before(minSupportedDate) || d.After(maxSupportedDate) {
		return ergo.Wrap(ErrOutOfSupportedRange, fmt.Sprintf("date=%v, min=%v, max=%v", t, minSupportedDate, maxSupportedDate))
	}
	return nil
}

func (c *JapaneseLunisolarCalendar) checkLunarYearRange(year int) error {
	if year < minLunisolarYear || year > maxLunisolarYear {
		return ergo.Wrap(ErrOutOfLunisolarRange, fmt.Sprintf("year=%d, min=%d, max=%d", year, minLunisolarYear, maxLunisolarYear))
	}
	return nil
}

// getYearInfo は年の暦情報 [閏月, 正月月, 正日, 日数パターン] を返します
func (c *JapaneseLunisolarCalendar) getYearInfo(lunarYear int, index int) (int, error) {
	if err := c.checkLunarYearRange(lunarYear); err != nil {
		return 0, err
	}
	return yearInfo[lunarYear-minLunisolarYear][index], nil
}

// GetYearInfo は指定した旧暦年の暦情報を返します。index の意味は次のとおりです:
// 0=閏月（なければ 0。閏6月なら 6）、1=旧正月の月、2=旧正月の日、
// 3=各月の日数パターン（最上位ビットから第1月・第2月…の順に 1=30日/0=29日の16ビット値）。
// 範囲外の年には ErrOutOfLunisolarRange をラップしたエラーを返します。
func (c *JapaneseLunisolarCalendar) GetYearInfo(lunarYear int, index int) (int, error) {
	return c.getYearInfo(lunarYear, index)
}

// internalGetDaysInMonth は太陰太陽暦の指定年/月の日数を返します
func (c *JapaneseLunisolarCalendar) internalGetDaysInMonth(lunarYear, lunarMonth int) (int, error) {
	daysPattern, err := c.getYearInfo(lunarYear, 3)
	if err != nil {
		return 0, err
	}
	// 16ビットマスクの上位から月を探索 (1なら30日、0なら29日)
	if (daysPattern>>(16-lunarMonth))&1 == 1 {
		return 30, nil
	}
	return 29, nil
}

// gregorianToLunar はグレゴリオ暦を太陰太陽暦に変換します
func (c *JapaneseLunisolarCalendar) gregorianToLunar(t time.Time) (lunarYear, lunarMonth, lunarDay int, err error) {
	if err = c.checkDateRange(t); err != nil {
		return 0, 0, 0, err
	}

	solarYear, solarMonth, solarDate := t.Date()

	// 太陽暦の年初からの通算日を計算
	dayOfYear := t.YearDay()

	lunarYear = solarYear
	if lunarYear == maxLunisolarYear+1 {
		// サポート最終旧暦年の年末はグレゴリオ暦の翌年 1 月にかかる
		// (旧暦 2049 年の大晦日 = 2050-01-22)。範囲チェック済みのこの日付は
		// 必ず前の旧暦年に属する (.NET の実装と同じ特例処理)。
		lunarYear--
		dayOfYear += daysInGregorianYear(lunarYear)
	} else {
		jan1Month, jErr := c.getYearInfo(lunarYear, 1)
		if jErr != nil {
			return 0, 0, 0, jErr
		}
		jan1Date, jErr := c.getYearInfo(lunarYear, 2)
		if jErr != nil {
			return 0, 0, 0, jErr
		}
		// 指定日が旧暦の前年に属するか判定
		if solarMonth < time.Month(jan1Month) || (solarMonth == time.Month(jan1Month) && solarDate < jan1Date) {
			lunarYear--
			dayOfYear += daysInGregorianYear(lunarYear)
		}
	}

	// 旧暦の元旦からの通算日を計算
	jan1Month, err := c.getYearInfo(lunarYear, 1)
	if err != nil {
		return 0, 0, 0, err
	}
	jan1Date, err := c.getYearInfo(lunarYear, 2)
	if err != nil {
		return 0, 0, 0, err
	}
	lunarDayOfYear := dayOfYear - (time.Date(lunarYear, time.Month(jan1Month), jan1Date, 0, 0, 0, 0, time.UTC).YearDay() - 1)

	// 通算日から月と日を計算
	// その年の月数 (12 or 13) を上限として走査する。
	// 検証済みテーブルでは通算日が年内日数を超えることはないが、万一テーブルが
	// 破損した場合にも存在しない月 (14月など) を返さないための防御として、
	// 超過時はサポート範囲外エラーを返す。
	monthsInYear, err := c.GetMonthsInYear(lunarYear)
	if err != nil {
		return 0, 0, 0, err
	}
	lunarDay = lunarDayOfYear
	for lunarMonth = 1; lunarMonth <= monthsInYear; lunarMonth++ {
		daysInMonth, dErr := c.internalGetDaysInMonth(lunarYear, lunarMonth)
		if dErr != nil {
			return 0, 0, 0, dErr
		}
		if lunarDay <= daysInMonth {
			return lunarYear, lunarMonth, lunarDay, nil
		}
		lunarDay -= daysInMonth
	}

	// ここに到達した場合、暦テーブルの不整合により日付が年内に収まらない
	return 0, 0, 0, ergo.Wrap(ErrOutOfSupportedRange,
		fmt.Sprintf("date=%v could not be mapped within lunar year=%d (months=%d)", t, lunarYear, monthsInYear))
}

// lunarToGregorian は太陰太陽暦をグレゴリオ暦に変換します。
// 月が年の月数を超える場合は ErrInvalidMonth を、日が月の日数を超える場合は
// ErrInvalidDay をラップしたエラーを返します。
func (c *JapaneseLunisolarCalendar) lunarToGregorian(lunarYear, lunarMonth, lunarDay int) (time.Time, error) {
	if err := c.checkLunarYearRange(lunarYear); err != nil {
		return time.Time{}, err
	}

	monthsInYear, err := c.GetMonthsInYear(lunarYear)
	if err != nil {
		return time.Time{}, err
	}
	if lunarMonth < 1 || lunarMonth > monthsInYear {
		return time.Time{}, ergo.Wrap(ErrInvalidMonth, fmt.Sprintf("month=%d, year=%d", lunarMonth, lunarYear))
	}
	daysInMonth, err := c.internalGetDaysInMonth(lunarYear, lunarMonth)
	if err != nil {
		return time.Time{}, err
	}
	if lunarDay < 1 || lunarDay > daysInMonth {
		return time.Time{}, ergo.Wrap(ErrInvalidDay, fmt.Sprintf("day=%d, month=%d, year=%d", lunarDay, lunarMonth, lunarYear))
	}

	// 旧暦の元旦からの通算日数を計算
	dayOfYear := lunarDay - 1
	for m := 1; m < lunarMonth; m++ {
		days, _ := c.internalGetDaysInMonth(lunarYear, m)
		dayOfYear += days
	}

	// グレゴリオ暦の旧暦元旦の日付を取得
	jan1Month, _ := c.getYearInfo(lunarYear, 1)
	jan1Date, _ := c.getYearInfo(lunarYear, 2)
	gregorianStartDate := time.Date(lunarYear, time.Month(jan1Month), jan1Date, 0, 0, 0, 0, time.UTC)

	// 日数を加算して日付を計算
	gregorianDate := gregorianStartDate.AddDate(0, 0, dayOfYear)
	return gregorianDate, nil
}

// --- 公開メソッド ---

// GetEra は指定されたグレゴリオ暦の日付に対応する元号IDを返します。
// 判定は t の壁時計上の日付（タイムゾーンに依存しない年月日）で行います。
// サポート範囲外の日付には ErrOutOfSupportedRange をラップしたエラーを返します。
func (c *JapaneseLunisolarCalendar) GetEra(t time.Time) (int, error) {
	if err := c.checkDateRange(t); err != nil {
		return 0, err
	}
	// erasは新しい順なので、最初に見つかったものが正解
	d := toUTCDate(t)
	for _, era := range c.eras {
		if !d.Before(era.StartDate) {
			return era.Era, nil
		}
	}
	return 0, ErrEraNotFound
}

// Eras はサポートされている元号のリストを返します
func (c *JapaneseLunisolarCalendar) Eras() []int {
	eraNumbers := make([]int, len(c.eras))
	for i, era := range c.eras {
		eraNumbers[i] = era.Era
	}
	sort.Ints(eraNumbers) // 昇順で返す
	return eraNumbers
}

// GetYear はグレゴリオ暦の日付が属する旧暦年を、元号年として返します。
// 例えば 2020-01-20 は旧暦 2019 年（同年の旧正月 2020-01-25 より前）に属するため、
// 令和元年として 1 を返します。GetMonth / GetDayOfMonth と組み合わせた値は
// ToDateTime で元のグレゴリオ暦日付に逆変換できます（.NET の
// JapaneseLunisolarCalendar.GetYear と同じ規則）。
//
// 注意: 元号は t の日付そのもので判定するため、改元日から次の旧正月までの期間
// （平成改元では 1989-01-08〜1989-02-05）は旧暦年が前の元号に属し、結果が
// 0 になります（平成 0 年 = 昭和 63 年に相当）。これも .NET と同じ挙動です。
func (c *JapaneseLunisolarCalendar) GetYear(t time.Time) (int, error) {
	lunarYear, _, _, err := c.gregorianToLunar(t)
	if err != nil {
		return 0, err
	}

	eraID, err := c.GetEra(t)
	if err != nil {
		return 0, err
	}

	for _, era := range c.eras {
		if era.Era == eraID {
			return lunarYear - era.YearOffset, nil
		}
	}
	return 0, ErrInternalEraNotFound
}

// GetMonth はグレゴリオ暦の日付に対応する旧暦の月を返します。
// 戻り値は年初からの通し番号（1〜12、閏月のある年は 1〜13）であり、
// 閏月がある年では閏月以降の値が伝統的な月名と 1 ずれます
// （例: 閏6月のある年では 7 が「閏6月」を指します）。
// 伝統的な月番号と閏月フラグが必要な場合は GetLunarDate を使用してください。
func (c *JapaneseLunisolarCalendar) GetMonth(t time.Time) (int, error) {
	_, month, _, err := c.gregorianToLunar(t)
	return month, err
}

// GetDayOfMonth はグレゴリオ暦の日付に対応する旧暦の日を返します
func (c *JapaneseLunisolarCalendar) GetDayOfMonth(t time.Time) (int, error) {
	_, _, day, err := c.gregorianToLunar(t)
	return day, err
}

// GetLunarDate はグレゴリオ暦の日付に対応する旧暦の日付を、伝統的な月番号で返します。
// month は 1〜12 の伝統的な月番号、isLeapMonth はその月が閏月かどうかを表します
// （例: 閏6月は month=6, isLeapMonth=true）。year は旧暦年の西暦表記です。
func (c *JapaneseLunisolarCalendar) GetLunarDate(t time.Time) (year, month, day int, isLeapMonth bool, err error) {
	year, month, day, err = c.gregorianToLunar(t)
	if err != nil {
		return 0, 0, 0, false, err
	}
	leapMonth, err := c.GetLeapMonth(year)
	if err != nil {
		return 0, 0, 0, false, err
	}
	if leapMonth > 0 && month > leapMonth {
		// 閏月の挿入位置以降は通し番号が伝統的な月番号より 1 大きい
		month--
		isLeapMonth = month == leapMonth
	}
	return year, month, day, isLeapMonth, nil
}

// IsLeapYear は指定された旧暦年（西暦表記）に閏月があるかどうかを返します。
// グレゴリオ暦の閏年判定ではない点に注意してください。
func (c *JapaneseLunisolarCalendar) IsLeapYear(lunarYear int) (bool, error) {
	leapMonth, err := c.getYearInfo(lunarYear, 0)
	if err != nil {
		return false, err
	}
	return leapMonth > 0, nil
}

// GetMonthsInYear は指定された年の月数を返します (12 or 13)
func (c *JapaneseLunisolarCalendar) GetMonthsInYear(lunarYear int) (int, error) {
	isLeap, err := c.IsLeapYear(lunarYear)
	if err != nil {
		return 0, err
	}
	if isLeap {
		return 13, nil
	}
	return 12, nil
}

// GetDaysInMonth は指定された旧暦の年/月の日数を返します
func (c *JapaneseLunisolarCalendar) GetDaysInMonth(lunarYear, lunarMonth int) (int, error) {
	if err := c.checkLunarYearRange(lunarYear); err != nil {
		return 0, err
	}
	// 月の妥当性チェック
	monthsInYear, _ := c.GetMonthsInYear(lunarYear)
	if lunarMonth < 1 || lunarMonth > monthsInYear {
		return 0, ergo.Wrap(ErrInvalidMonth, fmt.Sprintf("month=%d, year=%d", lunarMonth, lunarYear))
	}

	return c.internalGetDaysInMonth(lunarYear, lunarMonth)
}

// ToDateTime は旧暦の和暦表記（元号年、月、日）をグレゴリオ暦の time.Time に変換します。
// month は GetMonth と同じ年初からの通し番号（1〜12、閏月のある年は 1〜13）です。
// 戻り値は UTC の 0 時の time.Time です。
func (c *JapaneseLunisolarCalendar) ToDateTime(eraYear, month, day int, eraID int) (time.Time, error) {
	var lunarYear int
	var foundEra bool
	for _, era := range c.eras {
		if era.Era == eraID {
			lunarYear = eraYear + era.YearOffset
			foundEra = true
			break
		}
	}
	if !foundEra {
		return time.Time{}, ergo.Wrap(ErrInvalidEra, fmt.Sprintf("eraID=%d", eraID))
	}

	t, err := c.lunarToGregorian(lunarYear, month, day)
	if err != nil {
		return time.Time{}, err
	}

	// 変換後の日付がサポート範囲内か最終チェック
	if err := c.checkDateRange(t); err != nil {
		return time.Time{}, err
	}

	return t, nil
}

// GetGregorianYear は元号年と元号IDからグレゴリオ暦年を返します
func (c *JapaneseLunisolarCalendar) GetGregorianYear(eraYear, eraID int) (int, error) {
	for _, era := range c.eras {
		if era.Era == eraID {
			return era.YearOffset + eraYear, nil
		}
	}
	return 0, ergo.Wrap(ErrInvalidEra, fmt.Sprintf("eraID=%d", eraID))
}

// GetLeapMonth は指定された旧暦年（西暦表記）の閏月の伝統的な月番号を返します。
// 閏6月のある年は 6 を返し、閏月のない年は 0 を返します。
// （.NET の GetLeapMonth が返す「年初からの通し番号」(この例では 7) とは異なります。）
func (c *JapaneseLunisolarCalendar) GetLeapMonth(lunarYear int) (int, error) {
	return c.getYearInfo(lunarYear, 0)
}
