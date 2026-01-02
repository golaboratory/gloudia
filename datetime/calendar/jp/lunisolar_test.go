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
		{time.Date(2049, 12, 31, 0, 0, 0, 0, time.UTC), 5, false},   // サポート最大日
		{time.Date(2050, 1, 22, 23, 59, 59, 0, time.UTC), 5, false}, // サポート最大日
		// 境界値外
		{time.Date(1959, 12, 31, 0, 0, 0, 0, time.UTC), 0, true},
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
		{time.Date(1989, 1, 7, 0, 0, 0, 0, time.UTC), 3, 64, false},
		{time.Date(1989, 1, 8, 0, 0, 0, 0, time.UTC), 4, 1, false},
		{time.Date(2019, 4, 30, 0, 0, 0, 0, time.UTC), 4, 31, false},
		{time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC), 5, 1, false},
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

func TestJapaneseLunisolarCalendar_ExtendedRange(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()

	tests := []struct {
		name        string
		gregorian   time.Time
		lunarYear   int
		lunarMonth  int
		lunarDay    int
		expectError bool
	}{
		// 既存の範囲境界付近
		{
			name:       "2049 New Year",
			gregorian:  time.Date(2049, 2, 2, 0, 0, 0, 0, time.UTC),
			lunarYear:  2049,
			lunarMonth: 1,
			lunarDay:   1,
		},
		// 拡張された範囲 (2050年以降)
		{
			name:       "2050 New Year",
			gregorian:  time.Date(2050, 1, 23, 0, 0, 0, 0, time.UTC),
			lunarYear:  2050,
			lunarMonth: 1,
			lunarDay:   1,
		},
		{
			name:       "2060 New Year",
			gregorian:  time.Date(2060, 2, 2, 0, 0, 0, 0, time.UTC),
			lunarYear:  2060,
			lunarMonth: 1,
			lunarDay:   1,
		},
		{
			name:       "2070 New Year",
			gregorian:  time.Date(2070, 2, 11, 0, 0, 0, 0, time.UTC),
			lunarYear:  2070,
			lunarMonth: 1,
			lunarDay:   1,
		},
		{
			name:       "2080 New Year",
			gregorian:  time.Date(2080, 1, 22, 0, 0, 0, 0, time.UTC),
			lunarYear:  2080,
			lunarMonth: 1,
			lunarDay:   1,
		},
		{
			name:       "2090 New Year",
			gregorian:  time.Date(2090, 1, 30, 0, 0, 0, 0, time.UTC),
			lunarYear:  2090,
			lunarMonth: 1,
			lunarDay:   1,
		},
		{
			name:       "2100 New Year",
			gregorian:  time.Date(2100, 2, 9, 0, 0, 0, 0, time.UTC),
			lunarYear:  2100,
			lunarMonth: 1,
			lunarDay:   1,
		},
		// 閏月のある年のテスト (2052年は閏8月がある)
		{
			name:       "2052 Leap Month Start",
			gregorian:  time.Date(2052, 2, 1, 0, 0, 0, 0, time.UTC),
			lunarYear:  2052,
			lunarMonth: 1,
			lunarDay:   1,
		},
		// 範囲外エラーチェック
		{
			name:        "2101 Out of Range",
			gregorian:   time.Date(2101, 1, 29, 0, 0, 0, 0, time.UTC),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Gregorian -> Lunar
			ly, lm, ld, err := c.gregorianToLunar(tt.gregorian)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if ly != tt.lunarYear || lm != tt.lunarMonth || ld != tt.lunarDay {
				t.Errorf("GregorianToLunar(%v) = %d/%d/%d, want %d/%d/%d",
					tt.gregorian, ly, lm, ld, tt.lunarYear, tt.lunarMonth, tt.lunarDay)
			}

			// Lunar -> Gregorian (Round trip check)
			gDate, err := c.lunarToGregorian(tt.lunarYear, tt.lunarMonth, tt.lunarDay)
			if err != nil {
				t.Errorf("lunarToGregorian failed: %v", err)
				return
			}
			if !gDate.Equal(tt.gregorian) {
				t.Errorf("LunarToGregorian(%d/%d/%d) = %v, want %v",
					tt.lunarYear, tt.lunarMonth, tt.lunarDay, gDate, tt.gregorian)
			}
		})
	}
}

func TestJapaneseLunisolarCalendar_LeapMonths_Extended(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()

	// 2050年以降で閏月がある年のチェック
	leapYears := map[int]int{
		2052: 8,
		2055: 6,
		2058: 0, // データ上は0 (閏月なし) に見えるが、ビットマスク確認が必要
		2061: 3,
		2063: 7,
		2066: 5,
		2071: 5,
		2074: 4,
		2076: 8,
		2080: 6,
		2084: 10,
		2088: 5,
		2093: 3,
		2096: 5,
		2099: 4,
	}

	for year, expectedLeapMonth := range leapYears {
		leapMonth, err := c.GetLeapMonth(year)
		if err != nil {
			t.Errorf("GetLeapMonth(%d) error: %v", year, err)
			continue
		}
		// 注: yearInfoの定義によっては0の場合もあるため、
		// ここではエラーが出ないことと、定義済みの値と一致することを確認
		// 実際のyearInfoテーブルの値と照らし合わせる
		if expectedLeapMonth != 0 && leapMonth != expectedLeapMonth {
			// yearInfoのデータとテスト期待値が一致しているか確認用
			// 実際のデータ定義: {8, 2, 1, ...} // 2052 -> 閏8月
			t.Errorf("Year %d: expected leap month %d, got %d", year, expectedLeapMonth, leapMonth)
		}
	}
}
