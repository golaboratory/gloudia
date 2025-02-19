package ref

import (
	"fmt"
	"testing"
)

func dummyFunction() {}

type testStruct struct{}

func (t *testStruct) methodFunction() {}

func TestGetFuncNameHappyPath(t *testing.T) {
	tests := []struct {
		name     string
		fn       interface{}
		expected string
	}{
		{
			name:     "SimpleFunction",
			fn:       dummyFunction,
			expected: "dummyFunction",
		},
		{
			name:     "MethodFunction",
			fn:       (&testStruct{}).methodFunction,
			expected: "methodFunction",
		},
		// 匿名関数の名称は不定のため、検証から除外
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFuncName(tt.fn)
			if err != nil {
				t.Errorf("GetFuncName() error = %v", err)
				return
			}
			fmt.Println(got)
			if got != tt.expected {
				t.Errorf("GetFuncName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetFuncNameEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "NonFunction",
			input:   "not a function",
			wantErr: true,
		},
		{
			name:    "NilInput",
			input:   nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetFuncName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFuncName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetStructName(t *testing.T) {
	type Dummy struct{}
	tests := []struct {
		name     string
		input    any
		expected string
		wantErr  bool
	}{
		{
			name:     "StructValue",
			input:    Dummy{},
			expected: "Dummy",
			wantErr:  false,
		},
		{
			name:     "StructPointer",
			input:    &Dummy{},
			expected: "Dummy",
			wantErr:  false,
		},
		{
			name:    "NonStruct",
			input:   123,
			wantErr: true,
		},
		{
			name:    "NilInput",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetStructName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStructName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("GetStructName() = %v, want %v", got, tt.expected)
			}
		})
	}
}
