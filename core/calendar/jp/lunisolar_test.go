package jp

import (
	"testing"
	"time"
)

func TestGetYearInfo(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	tests := []struct {
		year  int
		index int
		want  int
	}{
		{1960, 0, 6},
		{1961, 0, 0},
		{1962, 0, 0},
		{1963, 0, 4},
		{1964, 0, 0},
		{2019, 0, 0},
		{2020, 0, 4},
		{2021, 0, 0},
		{2022, 0, 0},
		{2023, 0, 2},
	}
	for _, tt := range tests {
		got, err := c.GetYearInfo(tt.year, tt.index)
		if err != nil {
			t.Errorf("GetYearInfo(%d, %d) error: %v", tt.year, tt.index, err)
			continue
		}
		if got != tt.want {
			t.Errorf("GetYearInfo(%d, %d) = %d, want %d", tt.year, tt.index, got, tt.want)
		}
	}
}

func TestGetMonthsInYear(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	tests := []struct {
		year int
		want int
	}{
		{1960, 13},
		{1961, 12},
		{1963, 13},
		{1976, 13},
		{1984, 13},
		{1995, 13},
		{2006, 13},
		{2014, 13},
		{2023, 13},
		{2044, 13},
	}
	for _, tt := range tests {
		got, err := c.GetMonthsInYear(tt.year)
		if err != nil {
			t.Errorf("GetMonthsInYear(%d) error: %v", tt.year, err)
			continue
		}
		if got != tt.want {
			t.Errorf("GetMonthsInYear(%d) = %d, want %d", tt.year, got, tt.want)
		}
	}
}

func TestGetDaysInMonth(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	tests := []struct {
		year  int
		month int
		want  int
	}{
		{1960, 1, 30},
		{1960, 2, 29},
		{1960, 3, 30},
		{1960, 4, 29},
		{1960, 5, 30},
		{2019, 1, 30},
		{2019, 2, 29},
		{2019, 3, 30},
		{2019, 4, 29},
		{2019, 5, 30},
	}
	for _, tt := range tests {
		got, err := c.GetDaysInMonth(tt.year, tt.month)
		if err != nil {
			t.Errorf("GetDaysInMonth(%d, %d) error: %v", tt.year, tt.month, err)
			continue
		}
		if got != tt.want {
			t.Errorf("GetDaysInMonth(%d, %d) = %d, want %d", tt.year, tt.month, got, tt.want)
		}
	}
}

func TestGetEra(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	tests := []struct {
		date time.Time
		want int
	}{
		{time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC), 5}, // 令和開始日
		{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), 5},
		{time.Date(2018, 12, 31, 0, 0, 0, 0, time.UTC), 4}, // 平成
		{time.Date(1989, 1, 8, 0, 0, 0, 0, time.UTC), 4},   // 平成開始日
		{time.Date(1990, 6, 1, 0, 0, 0, 0, time.UTC), 4},
		{time.Date(1988, 12, 31, 0, 0, 0, 0, time.UTC), 3}, // 昭和
		{time.Date(1960, 1, 28, 0, 0, 0, 0, time.UTC), 3},
		{time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), 4},
		{time.Date(2049, 12, 31, 0, 0, 0, 0, time.UTC), 5},
		{time.Date(2019, 4, 30, 23, 59, 59, 0, time.UTC), 4}, // 平成最終日
	}
	for _, tt := range tests {
		got, err := c.GetEra(tt.date)
		if err != nil {
			t.Errorf("GetEra(%v) error: %v", tt.date, err)
			continue
		}
		if got != tt.want {
			t.Errorf("GetEra(%v) = %d, want %d", tt.date, got, tt.want)
		}
	}
}

func TestGetGregorianYear(t *testing.T) {
	c := NewJapaneseLunisolarCalendar()
	tests := []struct {
		eraYear int
		era     int
		want    int
	}{
		{1, 5, 2019},
		{2, 5, 2020},
		{5, 5, 2023},
		{1, 4, 1989},
		{10, 4, 1998},
		{31, 4, 2019},
		{1, 3, 1926},
		{35, 3, 1960},
		{64, 3, 1989},
		{30, 5, 2048},
	}
	for _, tt := range tests {
		got, err := c.GetGregorianYear(tt.eraYear, tt.era)
		if err != nil {
			t.Errorf("GetGregorianYear(%d, %d) error: %v", tt.eraYear, tt.era, err)
			continue
		}
		if got != tt.want {
			t.Errorf("GetGregorianYear(%d, %d) = %d, want %d", tt.eraYear, tt.era, got, tt.want)
		}
	}
}
