package core

import (
	"testing"
)

func TestClone(t *testing.T) {
	type SourceStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	type TargetStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	source := SourceStruct{Name: "test", Value: 123}
	expected := TargetStruct{Name: "test", Value: 123}

	result, err := Clone[SourceStruct, TargetStruct](source)
	if err != nil {
		t.Errorf("Clone returned an error: %v", err)
	}

	if result != expected {
		t.Errorf("Clone = %v; want %v", result, expected)
	}
}
