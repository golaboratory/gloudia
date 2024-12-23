package security

import (
	"testing"
)

func TestValidateStrength(t *testing.T) {
	tests := []struct {
		password      string
		includeUpper  bool
		includeLower  bool
		includeNumber bool
		includeSymbol bool
		minLength     int
		expected      bool
	}{
		{"Password123!", true, true, true, true, 8, true},
		{"password123", false, true, true, false, 8, true},
		{"PASSWORD123", true, false, true, false, 8, true},
		{"Password", true, true, false, false, 8, true},
		{"Pass123!", true, true, true, true, 8, true},
		{"password", false, true, false, false, 8, true},
		{"PASSWORD", true, false, false, false, 8, true},
		{"12345678", false, false, true, false, 8, true},
		{"!@#$%^&*", false, false, false, true, 8, true},
	}

	for _, tt := range tests {
		result, _ := ValidateStrength(tt.password, tt.includeUpper, tt.includeLower, tt.includeNumber, tt.includeSymbol, tt.minLength)
		if result != tt.expected {
			t.Errorf("ValidateStrength(%v, %v, %v, %v, %v, %v) = %v; want %v", tt.password, tt.includeUpper, tt.includeLower, tt.includeNumber, tt.includeSymbol, tt.minLength, result, tt.expected)
		}
	}
}

func TestGeneratePassword(t *testing.T) {
	tests := []struct {
		includeUpper  bool
		includeLower  bool
		includeNumber bool
		includeSymbol bool
		length        int
	}{
		{true, true, true, true, 12},
		{true, true, true, false, 8},
		{true, true, false, false, 10},
		{false, true, true, true, 15},
		{false, false, true, true, 6},
	}

	for _, tt := range tests {
		password := GeneratePassword(tt.includeUpper, tt.includeLower, tt.includeNumber, tt.includeSymbol, tt.length)
		if len(password) != tt.length {
			t.Errorf("GeneratePassword(%v, %v, %v, %v, %v) = %v; length = %v;", tt.includeUpper, tt.includeLower, tt.includeNumber, tt.includeSymbol, tt.length, len(password), tt.length)
		}

		if tt.includeUpper && !containsRunes(password, passwordUpperAlphabets) {
			t.Errorf("GeneratePassword(%v, %v, %v, %v, %v) = %v; missing upper case letters", tt.includeUpper, tt.includeLower, tt.includeNumber, tt.includeSymbol, tt.length, password)
		}

		if tt.includeLower && !containsRunes(password, passwordLowerAlphabets) {
			t.Errorf("GeneratePassword(%v, %v, %v, %v, %v) = %v; missing lower case letters", tt.includeUpper, tt.includeLower, tt.includeNumber, tt.includeSymbol, tt.length, password)
		}

		if tt.includeNumber && !containsRunes(password, passwordNumbers) {
			t.Errorf("GeneratePassword(%v, %v, %v, %v, %v) = %v; missing numbers", tt.includeUpper, tt.includeLower, tt.includeNumber, tt.includeSymbol, tt.length, password)
		}

		if tt.includeSymbol && !containsRunes(password, passwordSymbols) {
			t.Errorf("GeneratePassword(%v, %v, %v, %v, %v) = %v; missing symbols", tt.includeUpper, tt.includeLower, tt.includeNumber, tt.includeSymbol, tt.length, password)
		}
	}
}
