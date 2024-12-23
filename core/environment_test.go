package core

import (
	"os"
	"testing"
)

func TestGetStringVariable(t *testing.T) {
	os.Setenv("TEST_STRING", "value")
	defer os.Unsetenv("TEST_STRING")

	tests := []struct {
		key          string
		defaultValue string
		expected     string
	}{
		{"TEST_STRING", "default", "value"},
		{"NON_EXISTENT", "default", "default"},
	}

	for _, tt := range tests {
		result := GetStringVariable(tt.key, tt.defaultValue)
		if result != tt.expected {
			t.Errorf("GetStringVariable(%v, %v) = %v; want %v", tt.key, tt.defaultValue, result, tt.expected)
		}
	}
}

func TestGetIntVariable(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	tests := []struct {
		key          string
		defaultValue int
		expected     int
	}{
		{"TEST_INT", 0, 42},
		{"NON_EXISTENT", 10, 10},
		{"INVALID_INT", 10, 10},
	}

	for _, tt := range tests {
		if tt.key == "INVALID_INT" {
			os.Setenv("INVALID_INT", "invalid")
			defer os.Unsetenv("INVALID_INT")
		}
		result := GetIntVariable(tt.key, tt.defaultValue)
		if result != tt.expected {
			t.Errorf("GetIntVariable(%v, %v) = %v; want %v", tt.key, tt.defaultValue, result, tt.expected)
		}
	}
}

func TestGetBoolVariable(t *testing.T) {
	os.Setenv("TEST_BOOL", "1")
	defer os.Unsetenv("TEST_BOOL")

	tests := []struct {
		key          string
		defaultValue bool
		expected     bool
	}{
		{"TEST_BOOL", false, true},
		{"NON_EXISTENT", true, true},
		{"INVALID_BOOL", false, false},
	}

	for _, tt := range tests {
		if tt.key == "INVALID_BOOL" {
			os.Setenv("INVALID_BOOL", "invalid")
			defer os.Unsetenv("INVALID_BOOL")
		}
		result := GetBoolVariable(tt.key, tt.defaultValue)
		if result != tt.expected {
			t.Errorf("GetBoolVariable(%v, %v) = %v; want %v", tt.key, tt.defaultValue, result, tt.expected)
		}
	}
}
