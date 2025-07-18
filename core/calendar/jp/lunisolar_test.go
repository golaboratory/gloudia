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
		{time.Date(2050, 1, 23, 0, 0, 0, 0, time.UTC), 0, true},
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
		{time.Date(2050, 1, 23, 0, 0, 0, 0, time.UTC), 0, 0, true},
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
		{time.Date(2050, 1, 23, 0, 0, 0, 0, time.UTC), 0, true},
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
		{2050, false, true},
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
		{2050, 0, true},
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
