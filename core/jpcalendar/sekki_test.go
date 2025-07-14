package jpcalendar

import (
	"testing"
	"time"
)

func TestGregorianYearTo24SekkiList(t *testing.T) {
	type wantSekki struct {
		name  string
		month int
		day   int
	}
	cases := []struct {
		year int
		want []wantSekki
	}{
		{
			2024,
			[]wantSekki{
				{"小寒", 1, 6},
				{"大寒", 1, 20},
				{"立春", 2, 4},
				{"雨水", 2, 19},
				{"啓蟄", 3, 5},
				{"春分", 3, 20},
				{"清明", 4, 4},
				{"穀雨", 4, 19},
				{"立夏", 5, 5},
				{"小満", 5, 20},
				{"芒種", 6, 5},
				{"夏至", 6, 21},
				{"小暑", 7, 6},
				{"大暑", 7, 22},
				{"立秋", 8, 7},
				{"処暑", 8, 22},
				{"白露", 9, 7},
				{"秋分", 9, 22},
				{"寒露", 10, 8},
				{"霜降", 10, 23},
				{"立冬", 11, 7},
				{"小雪", 11, 22},
				{"大雪", 12, 7},
				{"冬至", 12, 21},
			},
		},

		{
			2023,
			[]wantSekki{
				{"小寒", 1, 6},
				{"大寒", 1, 20},
				{"立春", 2, 4},
				{"雨水", 2, 19},
				{"啓蟄", 3, 6},
				{"春分", 3, 21},
				{"清明", 4, 5},
				{"穀雨", 4, 20},
				{"立夏", 5, 6},
				{"小満", 5, 21},
				{"芒種", 6, 6},
				{"夏至", 6, 21},
				{"小暑", 7, 7},
				{"大暑", 7, 23},
				{"立秋", 8, 8},
				{"処暑", 8, 23},
				{"白露", 9, 8},
				{"秋分", 9, 23},
				{"寒露", 10, 8},
				{"霜降", 10, 24},
				{"立冬", 11, 8},
				{"小雪", 11, 22},
				{"大雪", 12, 7},
				{"冬至", 12, 22},
			},
		},
		{
			1900,
			[]wantSekki{
				{"小寒", 1, 6},
				{"春分", 3, 21},
				{"夏至", 6, 22},
				{"冬至", 12, 22},
			},
		},
		{
			2000,
			[]wantSekki{
				{"小寒", 1, 6},
				{"春分", 3, 20},
				{"夏至", 6, 21},
				{"冬至", 12, 21},
			},
		},
		{
			2010,
			[]wantSekki{
				{"小寒", 1, 5},
				{"春分", 3, 21},
				{"夏至", 6, 21},
				{"冬至", 12, 22},
			},
		},
		{
			2030,
			[]wantSekki{
				{"小寒", 1, 5},
				{"春分", 3, 20},
				{"夏至", 6, 21},
				{"冬至", 12, 22},
			},
		},
	}

	for _, c := range cases {
		got, err := GregorianYearTo24SekkiList(c.year)
		if err != nil {
			t.Errorf("year=%d: unexpected error: %v", c.year, err)
			continue
		}
		for _, w := range c.want {
			date := time.Date(c.year, time.Month(w.month), w.day, 0, 0, 0, 0, time.UTC)

			if wanted := got[date]; wanted != w.name {
				t.Errorf("year=%d: want %s at %v, got %s", c.year, w.name, date, got[date])
			}
		}
	}
}

func TestGregorianDateToSekki(t *testing.T) {
	cases := []struct {
		date time.Time
		want string
	}{
		{time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC), "小寒"},
		{time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC), "大寒"},
		{time.Date(2024, 2, 4, 0, 0, 0, 0, time.UTC), "立春"},
		{time.Date(2024, 2, 19, 0, 0, 0, 0, time.UTC), "雨水"},
		{time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC), "啓蟄"},
		{time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC), "春分"},
		{time.Date(2024, 4, 4, 0, 0, 0, 0, time.UTC), "清明"},
		{time.Date(2024, 4, 19, 0, 0, 0, 0, time.UTC), "穀雨"},
		{time.Date(2024, 5, 5, 0, 0, 0, 0, time.UTC), "立夏"},
		{time.Date(2024, 5, 20, 0, 0, 0, 0, time.UTC), "小満"},
		{time.Date(2024, 6, 5, 0, 0, 0, 0, time.UTC), "芒種"},
		{time.Date(2024, 6, 21, 0, 0, 0, 0, time.UTC), "夏至"},
	}
	for _, c := range cases {
		got, err := GregorianDateToSekki(c.date)
		if err != nil {
			t.Errorf("date=%v: unexpected error: %v", c.date, err)
		}
		if got != c.want {
			t.Errorf("date=%v: got %s, want %s", c.date, got, c.want)
		}
	}
}
