package jpcalendar

import (
	"testing"
	"time"
)

func TestGregorianMonthToWafuGetsumei(t *testing.T) {
	cases := []struct {
		month   int
		want    string
		wantErr bool
	}{
		{1, "睦月", false},
		{2, "如月", false},
		{3, "弥生", false},
		{4, "卯月", false},
		{5, "皐月", false},
		{6, "水無月", false},
		{7, "文月", false},
		{8, "葉月", false},
		{9, "長月", false},
		{10, "神無月", false},
		{11, "霜月", false},
		{12, "師走", false},
		{0, "", true},
		{13, "", true},
		{-1, "", true},
	}
	for _, c := range cases {
		got, err := GregorianMonthToWafuGetsumei(c.month)
		if (err != nil) != c.wantErr {
			t.Errorf("month=%d: unexpected error: %v", c.month, err)
		}
		if got != c.want {
			t.Errorf("month=%d: got %s, want %s", c.month, got, c.want)
		}
	}
}

func TestGregorianDateToWafuGetsumei(t *testing.T) {
	cases := []struct {
		date time.Time
		want string
	}{
		{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), "睦月"},
		{time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC), "如月"},
		{time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC), "弥生"},
		{time.Date(2024, 4, 10, 0, 0, 0, 0, time.UTC), "卯月"},
		{time.Date(2024, 5, 5, 0, 0, 0, 0, time.UTC), "皐月"},
		{time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC), "水無月"},
		{time.Date(2024, 7, 7, 0, 0, 0, 0, time.UTC), "文月"},
		{time.Date(2024, 8, 8, 0, 0, 0, 0, time.UTC), "葉月"},
		{time.Date(2024, 9, 9, 0, 0, 0, 0, time.UTC), "長月"},
		{time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC), "神無月"},
		{time.Date(2024, 11, 11, 0, 0, 0, 0, time.UTC), "霜月"},
		{time.Date(2024, 12, 12, 0, 0, 0, 0, time.UTC), "師走"},
	}
	for _, c := range cases {
		got, err := GregorianDateToWafuGetsumei(c.date)
		if err != nil {
			t.Errorf("date=%v: unexpected error: %v", c.date, err)
		}
		if got != c.want {
			t.Errorf("date=%v: got %s, want %s", c.date, got, c.want)
		}
	}
}
