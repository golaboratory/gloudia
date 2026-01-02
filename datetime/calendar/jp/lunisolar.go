package jp

import (
	"log/slog"
	"sort"
	"time"

	"github.com/newmo-oss/ergo"
)

// --- 定数とグローバル変数 ---

// サポートする和暦（太陰太陽暦）の年範囲
const (
	minLunisolarYear = 1960
	maxLunisolarYear = 2100
)

// サポートするグレゴリオ暦の日付範囲
var (
	minSupportedDate = time.Date(1960, 1, 28, 0, 0, 0, 0, time.UTC)
	maxSupportedDate = time.Date(2101, 1, 28, 23, 59, 59, 999999999, time.UTC)
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

// グレゴリオ暦の月までの通算日数
//var (
//	daysToMonth365 = []int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}
//	daysToMonth366 = []int{0, 31, 60, 91, 121, 152, 182, 213, 244, 274, 305, 335}
//)

// yearInfo は、C#の s_yinfo テーブルに対応します。
// 各行のデータ: [閏月(なければ0), 正月月, 正日, 各月の日数パターン(ビットマスク)]
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
	{3, 1, 23, 0b1110010101010000},  // 2050
	{0, 2, 11, 0b1001011011010000},  // 2051
	{8, 2, 1, 0b0100101011101000},   // 2052
	{0, 2, 19, 0b0100101011100000},  // 2053
	{0, 2, 8, 0b1010010011010000},   // 2054
	{6, 1, 28, 0b1101001001101000},  // 2055
	{0, 2, 15, 0b1101001001010000},  // 2056
	{0, 2, 4, 0b1101010100100000},   // 2057
	{0, 1, 24, 0b1101101010100000},  // 2058
	{0, 2, 12, 0b1011011010100000},  // 2059
	{0, 2, 2, 0b1001011011010000},   // 2060
	{3, 1, 21, 0b0100101011011000},  // 2061
	{0, 2, 9, 0b0100100110110000},   // 2062
	{7, 1, 29, 0b1010010010111000},  // 2063
	{0, 2, 17, 0b1010010010110000},  // 2064
	{0, 2, 5, 0b1011001001010000},   // 2065
	{5, 1, 26, 0b1011010100101000},  // 2066
	{0, 2, 14, 0b0110110101000000},  // 2067
	{0, 2, 3, 0b1010110110100000},   // 2068
	{0, 1, 23, 0b1001010110110000},  // 2069
	{0, 2, 11, 0b0100100110110000},  // 2070
	{5, 1, 31, 0b0110010010111000},  // 2071
	{0, 2, 19, 0b0110010010110000},  // 2072
	{0, 2, 7, 0b1101010010100000},   // 2073
	{4, 1, 27, 0b1110101001010000},  // 2074
	{0, 2, 15, 0b0110110101000000},  // 2075
	{8, 2, 5, 0b0101101011010000},   // 2076
	{0, 2, 23, 0b0010101101100000},  // 2077
	{0, 2, 12, 0b1001001101110000},  // 2078
	{0, 2, 2, 0b1001001011100000},   // 2079
	{6, 1, 22, 0b1100100101101000},  // 2080
	{0, 2, 9, 0b1100100101010000},   // 2081
	{0, 1, 29, 0b1101010010100000},  // 2082
	{0, 2, 17, 0b1101101001010000},  // 2083
	{10, 2, 6, 0b0101101010101000},  // 2084
	{0, 2, 24, 0b0101011011000000},  // 2085
	{0, 2, 13, 0b1010101011010000},  // 2086
	{0, 2, 3, 0b0010010111010000},   // 2087
	{5, 1, 24, 0b1001001011011000},  // 2088
	{0, 2, 10, 0b1100100101010000},  // 2089
	{0, 1, 30, 0b1010100101010000},  // 2090
	{0, 2, 18, 0b1011010010100000},  // 2091
	{0, 2, 7, 0b1011010101010000},   // 2092
	{3, 1, 27, 0b0101010110101000},  // 2093
	{0, 2, 15, 0b0100101110100000},  // 2094
	{0, 2, 5, 0b1010010110110000},   // 2095
	{5, 1, 25, 0b0101001010111000},  // 2096
	{0, 2, 12, 0b0101001010110000},  // 2097
	{0, 2, 1, 0b1010100101010000},   // 2098
	{4, 1, 21, 0b1011010010101000},  // 2099
	{0, 2, 9, 0b0110101010100000},   // 2100
}

// --- 構造体定義 ---

// EraInfo は元号の情報を保持します
type EraInfo struct {
	Era         int
	Name        string
	EnglishName string
	StartDate   time.Time
	YearOffset  int // グレゴリオ暦年から元号年を引いた値 (例: 令和 2019 - 1 = 2018)
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

func (c *JapaneseLunisolarCalendar) checkDateRange(t time.Time) error {
	if t.Before(minSupportedDate) || t.After(maxSupportedDate) {
		return ergo.New("out of the supported range", slog.Any("date", t), slog.Any("minSupportedDate", minSupportedDate), slog.Any("maxSupportedDate", maxSupportedDate))
	}
	return nil
}

func (c *JapaneseLunisolarCalendar) checkLunarYearRange(year int) error {
	if year < minLunisolarYear || year > maxLunisolarYear {
		return ergo.New("out of the supported lunisolar range", slog.Any("year", year), slog.Any("minLunisolarYear", minLunisolarYear), slog.Any("maxLunisolarYear", maxLunisolarYear))
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

// GetYearInfo は指定した年・インデックスの太陰太陽暦情報を返します（テスト用公開ラッパー）
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
	jan1Month, _ := c.getYearInfo(lunarYear, 1)
	jan1Date, _ := c.getYearInfo(lunarYear, 2)

	// 指定日が旧暦の前年に属するか判定
	if solarYear == lunarYear && (solarMonth < time.Month(jan1Month) || (solarMonth == time.Month(jan1Month) && solarDate < jan1Date)) {
		lunarYear--
		dayOfYear += daysInGregorianYear(lunarYear)
	}

	// 旧暦の元旦からの通算日を計算
	jan1Month, _ = c.getYearInfo(lunarYear, 1)
	jan1Date, _ = c.getYearInfo(lunarYear, 2)
	lunarDayOfYear := dayOfYear - (time.Date(lunarYear, time.Month(jan1Month), jan1Date, 0, 0, 0, 0, time.UTC).YearDay() - 1)

	// 通算日から月と日を計算
	lunarDay = lunarDayOfYear
	lunarMonth = 1
	for {
		daysInMonth, _ := c.internalGetDaysInMonth(lunarYear, lunarMonth)
		if lunarDay <= daysInMonth {
			break
		}
		lunarDay -= daysInMonth
		lunarMonth++
	}
	return lunarYear, lunarMonth, lunarDay, nil
}

// lunarToGregorian は太陰太陽暦をグレゴリオ暦に変換します
func (c *JapaneseLunisolarCalendar) lunarToGregorian(lunarYear, lunarMonth, lunarDay int) (time.Time, error) {
	if err := c.checkLunarYearRange(lunarYear); err != nil {
		return time.Time{}, err
	}

	// 旧暦の元旦からの通算日数を計算
	dayOfYear := lunarDay - 1
	for m := 1; m < lunarMonth; m++ {
		daysInMonth, _ := c.internalGetDaysInMonth(lunarYear, m)
		dayOfYear += daysInMonth
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

// GetEra は指定されたグレゴリオ暦の日付に対応する元号IDを返します
func (c *JapaneseLunisolarCalendar) GetEra(t time.Time) (int, error) {
	if err := c.checkDateRange(t); err != nil {
		return 0, err
	}
	// erasは新しい順なので、最初に見つかったものが正解
	for _, era := range c.eras {
		if !t.Before(era.StartDate) {
			return era.Era, nil
		}
	}
	return 0, ergo.New("could not determine era for the date")
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

// GetYear はグレゴリオ暦の日付に対応する和暦（元号年）を返します
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
			k := t.Year() - lunarYear
			return lunarYear + k - era.YearOffset, nil
		}
	}
	return 0, ergo.New("internal error: era not found")
}

// GetMonth はグレゴリオ暦の日付に対応する旧暦の月を返します
func (c *JapaneseLunisolarCalendar) GetMonth(t time.Time) (int, error) {
	_, month, _, err := c.gregorianToLunar(t)
	return month, err
}

// GetDayOfMonth はグレゴリオ暦の日付に対応する旧暦の日を返します
func (c *JapaneseLunisolarCalendar) GetDayOfMonth(t time.Time) (int, error) {
	_, _, day, err := c.gregorianToLunar(t)
	return day, err
}

// IsLeapYear は指定された年が閏年かどうかを返します
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
		return 0, ergo.New("invalid month for year", slog.Any("lunarMonth", lunarMonth), slog.Any("lunarYear", lunarYear))
	}

	return c.internalGetDaysInMonth(lunarYear, lunarMonth)
}

// ToDateTime は和暦（元号年、月、日）をグレゴリオ暦の time.Time に変換します
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
		return time.Time{}, ergo.New("invalid era", slog.Any("eraID", eraID))
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
	return 0, ergo.New("invalid era", slog.Any("eraID", eraID))
}

// GetLeapMonth は指定された年（西暦）の閏月を返します。
// 閏年でない場合は0を返します。
func (c *JapaneseLunisolarCalendar) GetLeapMonth(lunarYear int) (int, error) {
	// 指定された年がサポート範囲内か検証
	if err := c.checkLunarYearRange(lunarYear); err != nil {
		// エラーの場合は0とエラー情報を返す
		return 0, err
	}

	// yearInfoテーブルの0番目の要素（Leap Month情報）を取得
	leapMonth, err := c.getYearInfo(lunarYear, 0)
	if err != nil {
		// これは通常発生しないはずだが、念のため
		return 0, err
	}

	// 取得した値をそのまま返す（0の場合は閏月なし）
	return leapMonth, nil
}
