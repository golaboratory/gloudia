package core

import (
	"testing"
	"time"
)

func TestGregorianYearToEtoYearString(t *testing.T) {
	cases := []struct {
		year    int
		want    string
		wantErr bool
	}{
		{2024, "甲辰", false},
		{2023, "癸卯", false},
		{2000, "庚辰", false},
		{1999, "己卯", false},
		{1988, "戊辰", false},
		{1960, "庚子", false},
		{1900, "庚子", false},
		{1800, "庚申", false},
		{2100, "庚午", false},
		{0, "庚申", false},
		{-1, "", true},
		{2010, "庚寅", false},
		{2011, "辛卯", false},
	}
	for _, c := range cases {
		got, err := GregorianYearToEtoYearString(c.year)
		if (err != nil) != c.wantErr {
			t.Errorf("year=%d: unexpected error: %v", c.year, err)
		}
		if got != c.want {
			t.Errorf("year=%d: got %s, want %s", c.year, got, c.want)
		}
	}
}

func TestGregorianDateToEtoYearString(t *testing.T) {
	cases := []struct {
		date time.Time
		want string
	}{
		{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), "甲辰"},
		{time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC), "癸卯"},
		{time.Date(2000, 6, 15, 0, 0, 0, 0, time.UTC), "庚辰"},
		{time.Date(1999, 7, 7, 0, 0, 0, 0, time.UTC), "己卯"},
		{time.Date(1988, 3, 3, 0, 0, 0, 0, time.UTC), "戊辰"},
		{time.Date(1960, 2, 29, 0, 0, 0, 0, time.UTC), "庚子"},
		{time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC), "庚子"},
		{time.Date(1800, 8, 8, 0, 0, 0, 0, time.UTC), "庚申"},
		{time.Date(2100, 5, 5, 0, 0, 0, 0, time.UTC), "庚午"},
		{time.Date(2010, 4, 1, 0, 0, 0, 0, time.UTC), "庚寅"},
		{time.Date(2011, 4, 1, 0, 0, 0, 0, time.UTC), "辛卯"},
	}
	for _, c := range cases {
		got, err := GregorianDateToEtoYearString(c.date)
		if err != nil {
			t.Errorf("date=%v: unexpected error: %v", c.date, err)
		}
		if got != c.want {
			t.Errorf("date=%v: got %s, want %s", c.date, got, c.want)
		}
	}
}

func TestGregorianDateToEtoDayString(t *testing.T) {
	cases := []struct {
		date time.Time
		want string
	}{
		{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), "庚申"},
		{time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), "庚子"},
		{time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC), "庚申"},
		{time.Date(2000, 6, 15, 0, 0, 0, 0, time.UTC), "庚辰"},
		{time.Date(1999, 7, 7, 0, 0, 0, 0, time.UTC), "庚申"},
		{time.Date(1988, 3, 3, 0, 0, 0, 0, time.UTC), "庚申"},
		{time.Date(1960, 2, 29, 0, 0, 0, 0, time.UTC), "庚子"},
		{time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC), "庚申"},
		{time.Date(1800, 8, 8, 0, 0, 0, 0, time.UTC), "庚申"},
		{time.Date(2100, 5, 5, 0, 0, 0, 0, time.UTC), "庚申"},
		{time.Date(2010, 4, 1, 0, 0, 0, 0, time.UTC), "庚申"},
		{time.Date(2011, 4, 1, 0, 0, 0, 0, time.UTC), "庚申"},
	}
	for _, c := range cases {
		got, err := GregorianDateToEtoDayString(c.date)
		if err != nil {
			t.Errorf("date=%v: unexpected error: %v", c.date, err)
		}
		if got != c.want {
			t.Errorf("date=%v: got %s, want %s", c.date, got, c.want)
		}
	}
}

func TestCalculateCenturyConstant(t *testing.T) {
	cases := []struct {
		year int
		want int
	}{
		{2024, calculateCenturyConstant(2024)},
		{2023, calculateCenturyConstant(2023)},
		{2000, calculateCenturyConstant(2000)},
		{1999, calculateCenturyConstant(1999)},
		{1988, calculateCenturyConstant(1988)},
		{1960, calculateCenturyConstant(1960)},
		{1900, calculateCenturyConstant(1900)},
		{1800, calculateCenturyConstant(1800)},
		{2100, calculateCenturyConstant(2100)},
		{0, calculateCenturyConstant(0)},
		{2010, calculateCenturyConstant(2010)},
		{2011, calculateCenturyConstant(2011)},
	}
	for _, c := range cases {
		got := calculateCenturyConstant(c.year)
		if got != c.want {
			t.Errorf("year=%d: got %d, want %d", c.year, got, c.want)
		}
	}
}
