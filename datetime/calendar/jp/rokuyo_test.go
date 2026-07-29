package jp

import (
	"testing"
	"time"
)

func TestGregorianDateToRokuyoString(t *testing.T) {
	type testCase struct {
		date     time.Time
		expected string
		wantErr  bool
	}

	tests := []testCase{
		{time.Date(1960, 1, 28, 0, 0, 0, 0, time.UTC), "先勝", false},
		{time.Date(1960, 2, 1, 0, 0, 0, 0, time.UTC), "大安", false},
		{time.Date(1960, 6, 15, 0, 0, 0, 0, time.UTC), "友引", false},
		{time.Date(1960, 7, 18, 0, 0, 0, 0, time.UTC), "赤口", false},
		{time.Date(1960, 8, 1, 0, 0, 0, 0, time.UTC), "友引", false},
		{time.Date(1960, 8, 30, 0, 0, 0, 0, time.UTC), "先負", false},
		{time.Date(1960, 9, 1, 0, 0, 0, 0, time.UTC), "大安", false},
		{time.Date(1989, 1, 8, 0, 0, 0, 0, time.UTC), "赤口", false},
		{time.Date(1989, 2, 6, 0, 0, 0, 0, time.UTC), "先勝", false},
		{time.Date(1995, 8, 28, 0, 0, 0, 0, time.UTC), "仏滅", false},
		{time.Date(2006, 7, 25, 0, 0, 0, 0, time.UTC), "先勝", false},
		{time.Date(2014, 9, 24, 0, 0, 0, 0, time.UTC), "先負", false},
		{time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC), "大安", false},
		{time.Date(2023, 3, 22, 0, 0, 0, 0, time.UTC), "友引", false},
		{time.Date(2023, 4, 19, 0, 0, 0, 0, time.UTC), "赤口", false},
		{time.Date(2023, 4, 20, 0, 0, 0, 0, time.UTC), "先負", false},
		{time.Date(2044, 8, 1, 0, 0, 0, 0, time.UTC), "友引", false},
		{time.Date(2049, 12, 31, 0, 0, 0, 0, time.UTC), "赤口", false},
		{time.Date(1959, 12, 31, 0, 0, 0, 0, time.UTC), "", true},
		{time.Date(2200, 1, 23, 0, 0, 0, 0, time.UTC), "", true},
		{time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC), "", true},
		{time.Date(1961, 2, 15, 0, 0, 0, 0, time.UTC), "先勝", false},
		{time.Date(1976, 9, 1, 0, 0, 0, 0, time.UTC), "先負", false},
		{time.Date(1984, 11, 23, 0, 0, 0, 0, time.UTC), "仏滅", false},
		{time.Date(2011, 2, 3, 0, 0, 0, 0, time.UTC), "先勝", false},
		{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), "大安", false},
	}

	for i, tt := range tests {
		got, err := GregorianDateToRokuyoString(tt.date)
		if (err != nil) != tt.wantErr {
			t.Errorf("case %d: GregorianDateToRokuyoString(%v) error = %v, wantErr %v", i, tt.date, err, tt.wantErr)
			continue
		}
		if err == nil && got != tt.expected {
			t.Errorf("case %d: GregorianDateToRokuyoString(%v) = %v, want %v", i, tt.date, got, tt.expected)
		}
	}
}
