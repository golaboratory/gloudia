package jp

import (
	"testing"
	"time"
)

func TestGetEra(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	tests := []struct {
		date    time.Time
		want    int
		wantErr bool
	}{
		// 境界値・代表値
		{time.Date(1960, 1, 28, 0, 0, 0, 0, time.UTC), 3, false},    // 昭和
		{time.Date(1989, 1, 7, 23, 59, 59, 0, time.UTC), 3, false},  // 昭和最終日
		{time.Date(1989, 1, 8, 0, 0, 0, 0, time.UTC), 4, false},     // 平成開始日
		{time.Date(2019, 4, 30, 23, 59, 59, 0, time.UTC), 4, false}, // 平成最終日
		{time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC), 5, false},     // 令和開始日
		{time.Date(2049, 12, 31, 0, 0, 0, 0, time.UTC), 5, false},
		{time.Date(2050, 1, 22, 23, 59, 59, 0, time.UTC), 5, false}, // サポート最大日 (旧暦2049年大晦日)
		// 境界値外
		{time.Date(1959, 12, 31, 0, 0, 0, 0, time.UTC), 0, true},
		{time.Date(2050, 1, 23, 0, 0, 0, 0, time.UTC), 0, true}, // サポート最大日の翌日
		{time.Date(2200, 1, 23, 0, 0, 0, 0, time.UTC), 0, true},
		{time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC), 0, true},
		{time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC), 0, true},
	}
	for _, tt := range tests {
		got, err := c.GetEra(tt.date)
		if (err != nil) != tt.wantErr {
			t.Errorf("GetEra(%v) error = %v, wantErr %v", tt.date, err, tt.wantErr)
			continue
		}
		if err == nil && got != tt.want {
			t.Errorf("GetEra(%v) = %d, want %d", tt.date, got, tt.want)
		}
	}
}

func TestGetYearInfo(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	// サポート範囲内すべての年を網羅
	for year := minLunisolarYear; year <= maxLunisolarYear; year++ {
		for idx := 0; idx < 4; idx++ {
			_, err := c.GetYearInfo(year, idx)
			if err != nil {
				t.Errorf("GetYearInfo(%d, %d) unexpected error: %v", year, idx, err)
			}
		}
	}
	// 境界値外
	badYears := []int{minLunisolarYear - 1, maxLunisolarYear + 1, 0, 1800, 3000}
	for _, year := range badYears {
		_, err := c.GetYearInfo(year, 0)
		if err == nil {
			t.Errorf("GetYearInfo(%d, 0) expected error, got nil", year)
		}
	}
}

func TestGetMonthsInYear(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	for year := minLunisolarYear; year <= maxLunisolarYear; year++ {
		months, err := c.GetMonthsInYear(year)
		if err != nil {
			t.Errorf("GetMonthsInYear(%d) unexpected error: %v", year, err)
		}
		if months != 12 && months != 13 {
			t.Errorf("GetMonthsInYear(%d) = %d, want 12 or 13", year, months)
		}
	}
	// 境界値外
	badYears := []int{minLunisolarYear - 1, maxLunisolarYear + 1, 0, 1800, 3000}
	for _, year := range badYears {
		_, err := c.GetMonthsInYear(year)
		if err == nil {
			t.Errorf("GetMonthsInYear(%d) expected error, got nil", year)
		}
	}
}

func TestGetDaysInMonth(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	// サポート範囲内すべての年・月を網羅
	for year := minLunisolarYear; year <= maxLunisolarYear; year++ {
		months, _ := c.GetMonthsInYear(year)
		for month := 1; month <= months; month++ {
			days, err := c.GetDaysInMonth(year, month)
			if err != nil {
				t.Errorf("GetDaysInMonth(%d, %d) unexpected error: %v", year, month, err)
			}
			if days != 29 && days != 30 {
				t.Errorf("GetDaysInMonth(%d, %d) = %d, want 29 or 30", year, month, days)
			}
		}
	}
	// 境界値外
	type badCase struct{ year, month int }
	badCases := []badCase{
		{minLunisolarYear - 1, 1},
		{maxLunisolarYear + 1, 1},
		{1960, 0},
		{1960, 14},
		{2049, 0},
		{2049, 14},
		{0, 1},
		{3000, 1},
	}
	for _, bc := range badCases {
		_, err := c.GetDaysInMonth(bc.year, bc.month)
		if err == nil {
			t.Errorf("GetDaysInMonth(%d, %d) expected error, got nil", bc.year, bc.month)
		}
	}
}

func TestGetGregorianYear(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	type testCase struct {
		eraYear int
		eraID   int
		want    int
		wantErr bool
	}
	tests := []testCase{
		{1, 5, 2019, false},
		{2, 5, 2020, false},
		{1, 4, 1989, false},
		{10, 4, 1998, false},
		{35, 3, 1960, false},
		{64, 3, 1989, false},
		{31, 4, 2019, false},
		// 境界値外
		{1, 0, 0, true},
		{1, 99, 0, true},
		{0, 5, 2018, false},
		{-1, 5, 2017, false},
		{100, 5, 2118, false},
	}
	for _, tt := range tests {
		got, err := c.GetGregorianYear(tt.eraYear, tt.eraID)
		if (err != nil) != tt.wantErr {
			t.Errorf("GetGregorianYear(%d, %d) error = %v, wantErr %v", tt.eraYear, tt.eraID, err, tt.wantErr)
			continue
		}
		if err == nil && got != tt.want {
			t.Errorf("GetGregorianYear(%d, %d) = %d, want %d", tt.eraYear, tt.eraID, got, tt.want)
		}
	}
}

func TestGetYear(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	type testCase struct {
		date    time.Time
		wantEra int
		want    int
		wantErr bool
	}
	tests := []testCase{
		{time.Date(1960, 1, 28, 0, 0, 0, 0, time.UTC), 3, 35, false},
		// 1989-01-07 (昭和最終日) は旧暦では 1988 年 (昭和63年) 12月に属する
		{time.Date(1989, 1, 7, 0, 0, 0, 0, time.UTC), 3, 63, false},
		// 1989-01-08 (平成開始日) も旧暦では 1988 年に属するため、平成 0 年
		// (= 昭和 63 年相当) となる。.NET の JapaneseLunisolarCalendar と同じ挙動。
		{time.Date(1989, 1, 8, 0, 0, 0, 0, time.UTC), 4, 0, false},
		{time.Date(2019, 4, 30, 0, 0, 0, 0, time.UTC), 4, 31, false},
		{time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC), 5, 1, false},
		// 2020-01-20 は旧正月 (2020-01-25) より前なので旧暦 2019 年 = 令和元年
		{time.Date(2020, 1, 20, 0, 0, 0, 0, time.UTC), 5, 1, false},
		{time.Date(2049, 12, 31, 0, 0, 0, 0, time.UTC), 5, 31, false},
		// 境界値外
		{time.Date(1959, 12, 31, 0, 0, 0, 0, time.UTC), 0, 0, true},
		{time.Date(2200, 1, 23, 0, 0, 0, 0, time.UTC), 0, 0, true},
	}
	for _, tt := range tests {
		got, err := c.GetYear(tt.date)
		if (err != nil) != tt.wantErr {
			t.Errorf("GetYear(%v) error = %v, wantErr %v", tt.date, err, tt.wantErr)
			continue
		}
		if err == nil && got != tt.want {
			t.Errorf("GetYear(%v) = %d, want %d", tt.date, got, tt.want)
		}
	}
}

// TestGetYear_RoundTripsWithToDateTime は GetYear/GetMonth/GetDayOfMonth の組が
// ToDateTime で元の日付に戻ること (旧暦年ベースの一貫性) を検証する。
func TestGetYear_RoundTripsWithToDateTime(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	dates := []time.Time{
		time.Date(1960, 1, 28, 0, 0, 0, 0, time.UTC),
		time.Date(1989, 1, 8, 0, 0, 0, 0, time.UTC),  // 改元直後・旧正月前
		time.Date(2020, 1, 20, 0, 0, 0, 0, time.UTC), // 年初・旧正月前
		time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 7, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2050, 1, 22, 0, 0, 0, 0, time.UTC), // サポート最大日
	}
	for _, d := range dates {
		year, err1 := c.GetYear(d)
		month, err2 := c.GetMonth(d)
		day, err3 := c.GetDayOfMonth(d)
		era, err4 := c.GetEra(d)
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			t.Errorf("%v: unexpected error: %v %v %v %v", d, err1, err2, err3, err4)
			continue
		}
		back, err := c.ToDateTime(year, month, day, era)
		if err != nil {
			t.Errorf("ToDateTime(%d, %d, %d, %d) unexpected error: %v", year, month, day, era, err)
			continue
		}
		if !back.Equal(d) {
			t.Errorf("round trip failed: %v -> (%d, %d, %d, era=%d) -> %v", d, year, month, day, era, back)
		}
	}
}

func TestGetMonth(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	tests := []struct {
		date    time.Time
		want    int
		wantErr bool
	}{
		{time.Date(1960, 1, 28, 0, 0, 0, 0, time.UTC), 1, false},
		{time.Date(1960, 2, 15, 0, 0, 0, 0, time.UTC), 1, false},
		{time.Date(1989, 1, 8, 0, 0, 0, 0, time.UTC), 12, false},
		{time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC), 3, false},
		{time.Date(2049, 12, 31, 0, 0, 0, 0, time.UTC), 12, false},
		// 境界値外
		{time.Date(1959, 12, 31, 0, 0, 0, 0, time.UTC), 0, true},
	}
	for _, tt := range tests {
		got, err := c.GetMonth(tt.date)
		if (err != nil) != tt.wantErr {
			t.Errorf("GetMonth(%v) error = %v, wantErr %v", tt.date, err, tt.wantErr)
			continue
		}
		if err == nil && got != tt.want {
			t.Errorf("GetMonth(%v) = %d, want %d", tt.date, got, tt.want)
		}
	}
}

// TestGetMonth_SupportBoundary はサポート範囲の上限 (旧暦 2049 年大晦日 = 2050-01-22)
// の前後で正しく成功/失敗が切り替わることを検証する。
// かつて存在した 2050〜2100 年の暦テーブル拡張は、閏月の欠落等を含む誤データで
// あったため削除された (詳細は lunisolar.go の注記を参照)。
func TestGetMonth_SupportBoundary(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()

	// 2050-01-22 = 旧暦 2049年12月29日 (サポート最終日)
	month, err := c.GetMonth(time.Date(2050, 1, 22, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("GetMonth(2050-01-22) unexpected error: %v", err)
	}
	if month != 12 {
		t.Errorf("GetMonth(2050-01-22) = %d, want 12", month)
	}
	day, err := c.GetDayOfMonth(time.Date(2050, 1, 22, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("GetDayOfMonth(2050-01-22) unexpected error: %v", err)
	}
	if day != 29 {
		t.Errorf("GetDayOfMonth(2050-01-22) = %d, want 29", day)
	}

	// 翌日以降は範囲外エラー
	outOfRange := []time.Time{
		time.Date(2050, 1, 23, 0, 0, 0, 0, time.UTC),
		time.Date(2060, 6, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2100, 2, 9, 0, 0, 0, 0, time.UTC),
	}
	for _, d := range outOfRange {
		if _, err := c.GetMonth(d); err == nil {
			t.Errorf("GetMonth(%v) expected out-of-range error, got nil", d)
		}
	}
}

func TestGetDayOfMonth(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	tests := []struct {
		date    time.Time
		want    int
		wantErr bool
	}{
		{time.Date(1960, 1, 28, 0, 0, 0, 0, time.UTC), 1, false},
		{time.Date(1960, 2, 15, 0, 0, 0, 0, time.UTC), 19, false},
		{time.Date(1989, 1, 8, 0, 0, 0, 0, time.UTC), 1, false},
		{time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC), 27, false},
		{time.Date(2049, 12, 31, 0, 0, 0, 0, time.UTC), 7, false},
		// 境界値外
		{time.Date(2200, 1, 23, 0, 0, 0, 0, time.UTC), 0, true},
	}
	for _, tt := range tests {
		got, err := c.GetDayOfMonth(tt.date)
		if (err != nil) != tt.wantErr {
			t.Errorf("GetDayOfMonth(%v) error = %v, wantErr %v", tt.date, err, tt.wantErr)
			continue
		}
		if err == nil && got != tt.want {
			t.Errorf("GetDayOfMonth(%v) = %d, want %d", tt.date, got, tt.want)
		}
	}
}

func TestIsLeapYear(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	tests := []struct {
		year    int
		want    bool
		wantErr bool
	}{
		{1960, true, false},
		{1961, false, false},
		{1963, true, false},
		{1976, true, false},
		{1984, true, false},
		{1995, true, false},
		{2006, true, false},
		{2014, true, false},
		{2023, true, false},
		{2044, true, false},
		// 境界値外
		{1959, false, true},
		{2200, false, true},
		{0, false, true},
		{3000, false, true},
	}
	for _, tt := range tests {
		got, err := c.IsLeapYear(tt.year)
		if (err != nil) != tt.wantErr {
			t.Errorf("IsLeapYear(%d) error = %v, wantErr %v", tt.year, err, tt.wantErr)
			continue
		}
		if err == nil && got != tt.want {
			t.Errorf("IsLeapYear(%d) = %v, want %v", tt.year, got, tt.want)
		}
	}
}

func TestToDateTime(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	tests := []struct {
		eraYear int
		month   int
		day     int
		eraID   int
		want    time.Time
		wantErr bool
	}{
		{35, 1, 1, 3, time.Date(1960, 1, 28, 0, 0, 0, 0, time.UTC), false},
		{1, 1, 1, 4, time.Date(1989, 2, 6, 0, 0, 0, 0, time.UTC), false},
		{1, 4, 27, 5, time.Date(2019, 5, 31, 0, 0, 0, 0, time.UTC), false},
		{31, 11, 2, 5, time.Date(2049, 11, 26, 0, 0, 0, 0, time.UTC), false},
		// 境界値外
		{0, 1, 1, 3, time.Time{}, true},
		{1, 0, 1, 3, time.Time{}, true},
		{1, 1, 0, 3, time.Time{}, true},
		{1, 1, 1, 0, time.Time{}, true},
		{1, 14, 1, 3, time.Time{}, true},
		{1, 1, 32, 3, time.Time{}, true},
		// 年は範囲内だが月・日が不正なケース (令和2年 = 旧暦2020年は閏4月あり13ヶ月、1月は30日)
		{2, 14, 1, 5, time.Time{}, true},
		{2, 1, 31, 5, time.Time{}, true},
		{2, 1, 0, 5, time.Time{}, true},
	}
	for _, tt := range tests {
		got, err := c.ToDateTime(tt.eraYear, tt.month, tt.day, tt.eraID)
		if (err != nil) != tt.wantErr {
			t.Errorf("ToDateTime(%d, %d, %d, %d) error = %v, wantErr %v", tt.eraYear, tt.month, tt.day, tt.eraID, err, tt.wantErr)
			continue
		}
		if err == nil && !got.Equal(tt.want) {
			t.Errorf("ToDateTime(%d, %d, %d, %d) = %v, want %v", tt.eraYear, tt.month, tt.day, tt.eraID, got, tt.want)
		}
	}
}

func TestGetLeapMonth(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	type testCase struct {
		year    int
		want    int
		wantErr bool
	}
	// 1960〜2049年のうち閏月がある年とない年をいくつかピックアップ
	tests := []testCase{
		{1960, 6, false},  // 閏月あり
		{1961, 0, false},  // 閏月なし
		{1963, 4, false},  // 閏月あり
		{1976, 8, false},  // 閏月あり
		{1984, 10, false}, // 閏月あり
		{1995, 8, false},  // 閏月あり
		{2006, 7, false},  // 閏月あり
		{2014, 9, false},  // 閏月あり
		{2023, 2, false},  // 閏月あり
		{2025, 6, false},  // 閏月あり (閏6月)
		{2033, 11, false}, // 閏月あり (いわゆる旧暦2033年問題の年。閏11月を採用)
		{2044, 7, false},  // 閏月あり
		{1962, 0, false},  // 閏月なし
		{2049, 0, false},  // 閏月なし
		// 境界値外
		{1959, 0, true},
		{2200, 0, true},
		{0, 0, true},
		{3000, 0, true},
	}
	for _, tt := range tests {
		got, err := c.GetLeapMonth(tt.year)
		if (err != nil) != tt.wantErr {
			t.Errorf("GetLeapMonth(%d) error = %v, wantErr %v", tt.year, err, tt.wantErr)
			continue
		}
		if err == nil && got != tt.want {
			t.Errorf("GetLeapMonth(%d) = %d, want %d", tt.year, got, tt.want)
		}
	}
}

// TestGregorianToLunar_KnownDates は公表されている旧暦日付との一致を検証する
// (旧正月・閏月境界などの実データアンカー)。
func TestGregorianToLunar_KnownDates(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()

	tests := []struct {
		name       string
		gregorian  time.Time
		lunarYear  int
		lunarMonth int // 年初からの通し番号
		lunarDay   int
	}{
		{"1960 旧正月", time.Date(1960, 1, 28, 0, 0, 0, 0, time.UTC), 1960, 1, 1},
		{"2024 旧正月", time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC), 2024, 1, 1},
		{"2025 旧正月", time.Date(2025, 1, 29, 0, 0, 0, 0, time.UTC), 2025, 1, 1},
		{"2025 閏6月朔日 (通し7月)", time.Date(2025, 7, 25, 0, 0, 0, 0, time.UTC), 2025, 7, 1},
		{"2033 旧正月 (閏11月の年)", time.Date(2033, 1, 31, 0, 0, 0, 0, time.UTC), 2033, 1, 1},
		{"2049 旧正月", time.Date(2049, 2, 2, 0, 0, 0, 0, time.UTC), 2049, 1, 1},
		{"サポート最終日", time.Date(2050, 1, 22, 0, 0, 0, 0, time.UTC), 2049, 12, 29},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ly, lm, ld, err := c.gregorianToLunar(tt.gregorian)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ly != tt.lunarYear || lm != tt.lunarMonth || ld != tt.lunarDay {
				t.Errorf("gregorianToLunar(%v) = %d/%d/%d, want %d/%d/%d",
					tt.gregorian, ly, lm, ld, tt.lunarYear, tt.lunarMonth, tt.lunarDay)
			}

			// 逆変換のラウンドトリップ
			gDate, err := c.lunarToGregorian(tt.lunarYear, tt.lunarMonth, tt.lunarDay)
			if err != nil {
				t.Fatalf("lunarToGregorian failed: %v", err)
			}
			if !gDate.Equal(tt.gregorian) {
				t.Errorf("lunarToGregorian(%d/%d/%d) = %v, want %v",
					tt.lunarYear, tt.lunarMonth, tt.lunarDay, gDate, tt.gregorian)
			}
		})
	}
}

// TestGetLunarDate は伝統的な月番号と閏月フラグへの変換を検証する。
// 2025 年は閏6月がある (通し番号 7 が閏6月)。
func TestGetLunarDate(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()

	tests := []struct {
		name        string
		gregorian   time.Time
		year        int
		month       int
		day         int
		isLeapMonth bool
	}{
		{"閏月のない月", time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC), 2024, 1, 1, false},
		{"閏月直前 (6月30日)", time.Date(2025, 7, 24, 0, 0, 0, 0, time.UTC), 2025, 6, 30, false},
		{"閏6月朔日", time.Date(2025, 7, 25, 0, 0, 0, 0, time.UTC), 2025, 6, 1, true},
		{"閏6月の翌月 (7月1日)", time.Date(2025, 8, 23, 0, 0, 0, 0, time.UTC), 2025, 7, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			year, month, day, isLeap, err := c.GetLunarDate(tt.gregorian)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if year != tt.year || month != tt.month || day != tt.day || isLeap != tt.isLeapMonth {
				t.Errorf("GetLunarDate(%v) = (%d, %d, %d, %v), want (%d, %d, %d, %v)",
					tt.gregorian, year, month, day, isLeap, tt.year, tt.month, tt.day, tt.isLeapMonth)
			}
		})
	}
}

// TestWallClockDateInterpretation は JST など UTC 以外のロケーションの time.Time が
// 壁時計上の日付として解釈されることを検証する。
func TestWallClockDateInterpretation(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	jst := time.FixedZone("JST", 9*60*60)

	// 2019-05-01 00:00 JST は UTC の瞬間では 2019-04-30 だが、日付として令和と判定されるべき
	era, err := c.GetEra(time.Date(2019, 5, 1, 0, 0, 0, 0, jst))
	if err != nil {
		t.Fatalf("GetEra(2019-05-01 JST) unexpected error: %v", err)
	}
	if era != 5 {
		t.Errorf("GetEra(2019-05-01 00:00 JST) = %d, want 5 (令和)", era)
	}

	// サポート最小日の 0 時 (JST) も日付として範囲内であるべき
	month, err := c.GetMonth(time.Date(1960, 1, 28, 0, 0, 0, 0, jst))
	if err != nil {
		t.Fatalf("GetMonth(1960-01-28 JST) unexpected error: %v", err)
	}
	if month != 1 {
		t.Errorf("GetMonth(1960-01-28 00:00 JST) = %d, want 1", month)
	}

	// UTC-11 のような西側のロケーションでも壁時計日付で解釈されるべき
	west := time.FixedZone("SST", -11*60*60)
	if _, err := c.GetMonth(time.Date(2050, 1, 23, 0, 0, 0, 0, west)); err == nil {
		t.Error("GetMonth(2050-01-23 00:00 UTC-11) expected out-of-range error, got nil")
	}
}
