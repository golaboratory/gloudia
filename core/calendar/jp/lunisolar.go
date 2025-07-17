package jp

import (
	"fmt"
	"sort"
	"time"
)

// minLunisolarYear, maxLunisolarYear はサポートされる太陰太陽暦の最小・最大年です。
const (
	minLunisolarYear = 1960
	maxLunisolarYear = 2049
)

// minSupportedDate, maxSupportedDate は和暦太陰太陽暦でサポートされる日付範囲です。
var (
	minSupportedDate = time.Date(1960, 1, 28, 0, 0, 0, 0, time.UTC)
	maxSupportedDate = time.Date(2050, 1, 22, 23, 59, 59, 999999999, time.UTC)
)

// yearInfo は各年の太陰太陽暦情報を格納したテーブルです。
// 各要素は [閏月(なければ0), 正月月, 正日, 各月の日数パターン(ビットマスク)] です。
var yearInfo = [][4]int{
	/*Y       LM  M   D      DaysPerMonthPattern */
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

// EraInfo は元号の情報を表します。
type EraInfo struct {
	Era         int       // 元号番号
	Name        string    // 元号名（日本語）
	EnglishName string    // 元号名（英語）
	StartDate   time.Time // 元号の開始日
	YearOffset  int       // グレゴリオ暦年から元号年を引くためのオフセット
}

// getEraInfo はサポート範囲内の元号情報を新しい順で返します。
func getEraInfo() []EraInfo {
	// .NETの実装に合わせて元号データを定義
	// 実際にはもっと多くの元号があるが、カレンダーのサポート範囲(1960-)に関連するものに限定
	eras := []EraInfo{
		{5, "令和", "Reiwa", time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC), 2018},
		{4, "平成", "Heisei", time.Date(1989, 1, 8, 0, 0, 0, 0, time.UTC), 1988},
		{3, "昭和", "Showa", time.Date(1926, 12, 25, 0, 0, 0, 0, time.UTC), 1925},
		// {2, "大正", "Taisho", ...}, // 1960年より前なので除外
		// {1, "明治", "Meiji", ...},  // 1960年より前なので除外
	}
	return eras
}

// JapaneseLunisolarCalendar は日本の太陰太陽暦（和暦）を扱うカレンダー型です。
type JapaneseLunisolarCalendar struct {
	eras []EraInfo
}

// NewJapaneseLunisolarCalendar は JapaneseLunisolarCalendar の新しいインスタンスを生成します。
func NewJapaneseLunisolarCalendar() *JapaneseLunisolarCalendar {
	// C#のTrimErasに相当するロジックは、getEraInfoで元から絞ることで代替
	return &JapaneseLunisolarCalendar{
		eras: getEraInfo(),
	}
}

// MinSupportedDateTime はサポートされる最小日時を返します。
func (c *JapaneseLunisolarCalendar) MinSupportedDateTime() time.Time {
	return minSupportedDate
}

// MaxSupportedDateTime はサポートされる最大日時を返します。
func (c *JapaneseLunisolarCalendar) MaxSupportedDateTime() time.Time {
	return maxSupportedDate
}

// checkYear は指定した年がサポート範囲内かどうかを検証します。
func (c *JapaneseLunisolarCalendar) checkYear(lunarYear int) error {
	if lunarYear < minLunisolarYear || lunarYear > maxLunisolarYear {
		return fmt.Errorf("year %d is out of range. supported range is from %d to %d",
			lunarYear, minLunisolarYear, maxLunisolarYear)
	}
	return nil
}

// GetYearInfo は指定した年の太陰太陽暦情報を返します。
// index: 0=閏月, 1=正月月, 2=正日, 3=日数パターン
func (c *JapaneseLunisolarCalendar) GetYearInfo(lunarYear int, index int) (int, error) {
	if err := c.checkYear(lunarYear); err != nil {
		return 0, err
	}
	if index < 0 || index > 3 {
		return 0, fmt.Errorf("index out of range")
	}
	return yearInfo[lunarYear-minLunisolarYear][index], nil
}

// GetMonthsInYear は指定した年の月数（12または13）を返します。
func (c *JapaneseLunisolarCalendar) GetMonthsInYear(lunarYear int) (int, error) {
	leapMonth, err := c.GetYearInfo(lunarYear, 0)
	if err != nil {
		return 0, err
	}
	if leapMonth > 0 {
		return 13, nil
	}
	return 12, nil
}

// GetDaysInMonth は指定した年・月の日数（29または30）を返します。
func (c *JapaneseLunisolarCalendar) GetDaysInMonth(lunarYear, lunarMonth int) (int, error) {
	daysPattern, err := c.GetYearInfo(lunarYear, 3)
	if err != nil {
		return 0, err
	}

	// ビットマスクの15ビット目から月を数える
	// 1なら30日、0なら29日
	if (daysPattern>>(16-lunarMonth))&1 == 1 {
		return 30, nil
	}
	return 29, nil
}

// GetEra は指定したグレゴリオ暦日付に対応する元号番号を返します。
func (c *JapaneseLunisolarCalendar) GetEra(t time.Time) (int, error) {
	if t.Before(c.MinSupportedDateTime()) || t.After(c.MaxSupportedDateTime()) {
		return 0, fmt.Errorf("date is out of supported range")
	}
	// erasは新しい順なので、最初に見つかったものが正解
	for _, era := range c.eras {
		if !t.Before(era.StartDate) {
			return era.Era, nil
		}
	}
	return 0, fmt.Errorf("could not determine era for the date")
}

// Eras はサポートされている元号番号のリストを昇順で返します。
func (c *JapaneseLunisolarCalendar) Eras() []int {
	eraNumbers := make([]int, len(c.eras))
	for i, era := range c.eras {
		eraNumbers[i] = era.Era
	}
	// .NET実装は新しい順だが、ここでは昇順で返すのが一般的
	sort.Ints(eraNumbers)
	return eraNumbers
}

// GetGregorianYear は元号年と元号番号からグレゴリオ暦年を計算して返します。
func (c *JapaneseLunisolarCalendar) GetGregorianYear(eraYear, era int) (int, error) {
	for _, e := range c.eras {
		if e.Era == era {
			return eraYear + e.YearOffset, nil
		}
	}
	return 0, fmt.Errorf("invalid era: %d", era)
}
