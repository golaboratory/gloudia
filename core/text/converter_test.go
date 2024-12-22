package text

import (
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
	year := 2021
	expected := "辛丑" // Example expected value for the year 2021

	result, err := ConvertToZodiacByYear(year)
	if err != nil {
		t.Errorf("ConvertToZodiacByYear returned an error: %v", err)
	}

	if result != expected {
		t.Errorf("ConvertToZodiacByYear = %v; want %v", result, expected)
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
	dt := time.Date(2021, 5, 15, 0, 0, 0, 0, time.UTC)
	expected := "癸亥" // Replace with the actual expected value

	result, err := ConvertToDayZodiacByDate(dt)
	if err != nil {
		t.Errorf("ConvertToDayZodiacByDate returned an error: %v", err)
	}

	if result != expected {
		t.Errorf("ConvertToDayZodiacByDate = %v; want %v", result, expected)
	}
}
