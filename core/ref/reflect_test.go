package ref

import (
	"strings"
	"testing"
)

func dummyFunction() {}

type testStruct struct{}

func (t *testStruct) methodFunction() {}

func TestGetFuncNameHappyPath(t *testing.T) {
	tests := []struct {
		name     string
		fn       interface{}
		contains string
	}{
		{
			name:     "SimpleFunction",
			fn:       dummyFunction,
			contains: "dummyFunction",
		},
		{
			name:     "MethodFunction",
			fn:       (&testStruct{}).methodFunction,
			contains: "methodFunction",
		},
		{
			name:     "AnonymousFunction",
			fn:       func() {},
			contains: "func",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFuncName(tt.fn)
			if err != nil {
				t.Errorf("GetFuncName() error = %v", err)
				return
			}
			if !strings.Contains(got, tt.contains) {
				t.Errorf("GetFuncName() = %v, want containing %v", got, tt.contains)
			}
		})
	}
}

/*
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
*/
