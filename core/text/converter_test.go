package text

import (
	"fmt"
	"testing"
	"time"
)

func TestConvertToEraDate(t *testing.T) {
	dt := time.Date(2021, 5, 15, 0, 0, 0, 0, time.UTC)
	expected := "令和03年05月15日"

	result, err := ConvertToEraDate(dt)
	if err != nil {
		t.Errorf("ConvertToEraDate returned an error: %v", err)
	}

	if result != expected {
		t.Errorf("ConvertToEraDate = %v; want %v", result, expected)
	}
}

func TestConvertToEraDateWithFormat(t *testing.T) {
	dt := time.Date(2021, 5, 15, 14, 30, 0, 0, time.UTC)
	format := "ggggyy年MM月dd日 HH:mm:ss"
	expected := "令和03年05月15日 14:30:00"

	result, err := ConvertToEraDateWithFormat(dt, format)
	if err != nil {
		t.Errorf("ConvertToEraDateWithFormat returned an error: %v", err)
	}

	if result != expected {
		t.Errorf("ConvertToEraDateWithFormat = %v; want %v", result, expected)
	}
}

func TestConvertToZodiacByYear(t *testing.T) {
	tests := []struct {
		year     int
		expected string
	}{
		{2021, "辛丑"},
		{2020, "庚子"},
		{2019, "己亥"},
		{2025, "乙巳"},
		// 他のテストケースを追加できます
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Year %d", tt.year), func(t *testing.T) {
			result, err := ConvertToZodiacByYear(tt.year)
			if err != nil {
				t.Errorf("ConvertToZodiacByYear returned an error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("ConvertToZodiacByYear = %v; want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertToZodiacByDate(t *testing.T) {
	dt := time.Date(2021, 5, 15, 0, 0, 0, 0, time.UTC)
	expected := "辛丑" // Example expected value for the date 2021-05-15

	result, err := ConvertToZodiacByDate(dt)
	if err != nil {
		t.Errorf("ConvertToZodiacByDate returned an error: %v", err)
	}

	if result != expected {
		t.Errorf("ConvertToZodiacByDate = %v; want %v", result, expected)
	}
}

func TestConvertToDayZodiacByDate(t *testing.T) {
	tests := []struct {
		date     time.Time
		expected string
	}{
		{time.Date(2021, 5, 15, 0, 0, 0, 0, time.UTC), "癸亥"},
		{time.Date(2024, 12, 22, 0, 0, 0, 0, time.UTC), "庚申"},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Date %v", tt.date), func(t *testing.T) {
			result, err := ConvertToDayZodiacByDate(tt.date)
			if err != nil {
				t.Errorf("ConvertToDayZodiacByDate returned an error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("ConvertToDayZodiacByDate = %v; want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertCamelToKebab(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"CamelCaseString", "camel-case-string"},
		{"simpleTest", "simple-test"},
		{"ABC", "a-b-c"},
		{"already-kebab", "already-kebab"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Input %s", tt.input), func(t *testing.T) {
			got := ConvertCamelToKebab(tt.input)
			if got != tt.expected {
				t.Errorf("ConvertCamelToKebab(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}
