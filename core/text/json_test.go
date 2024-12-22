package text

import (
	"testing"
)

func TestSerializeJson(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	data := TestStruct{Name: "test", Value: 123}
	expected := `{"name":"test","value":123}`

	result, err := SerializeJson(data)
	if err != nil {
		t.Errorf("SerializeJson returned an error: %v", err)
	}

	if result != expected {
		t.Errorf("SerializeJson = %v; want %v", result, expected)
	}
}

func TestDeserializeJson(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	data := `{"name":"test","value":123}`
	expected := TestStruct{Name: "test", Value: 123}

	result, err := DeserializeJson[TestStruct](data)
	if err != nil {
		t.Errorf("DeserializeJson returned an error: %v", err)
	}

	if result != expected {
		t.Errorf("DeserializeJson = %v; want %v", result, expected)
	}
}
