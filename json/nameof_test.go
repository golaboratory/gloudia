package json_test

import (
	"testing"

	"github.com/golaboratory/gloudia/json"
)

type Child struct {
	Name string `json:"childName"`
	Age  int    `json:"childAge,omitempty"`
}

type Parent struct {
	ID        int    `json:"id"`
	Title     string `json:"title,omitempty"`
	NoTag     string
	ChildVal  Child  `json:"childVal"`
	ChildPtr  *Child `json:"childPtr"`
	NilPtr    *Child `json:"nilPtr"`
	Anonymous struct {
		Inner string `json:"inner"`
	} `json:"anonymous"`
}

func TestNameOf(t *testing.T) {
	// テストデータの準備
	child := Child{Name: "Taro", Age: 10}
	childPtr := &Child{Name: "Jiro", Age: 5}

	p := Parent{
		ID:       100,
		Title:    "Hello",
		NoTag:    "NoTagVal",
		ChildVal: child,
		ChildPtr: childPtr,
		NilPtr:   nil, // nil pointer
	}
	// Anonymous struct initialization
	p.Anonymous.Inner = "InnerVal"

	// 定義外の変数
	otherVar := 123

	tests := []struct {
		name          string
		root          any
		target        any
		expectedName  string
		expectedError bool
	}{
		{
			name:         "Simple field: ID",
			root:         &p,
			target:       &p.ID,
			expectedName: "id",
		},
		{
			name:         "Field with omitempty: Title",
			root:         &p,
			target:       &p.Title,
			expectedName: "title",
		},
		{
			name:         "Field without json tag: NoTag",
			root:         &p,
			target:       &p.NoTag,
			expectedName: "",
		},
		{
			name:         "Nested struct field: ChildVal.Name",
			root:         &p,
			target:       &p.ChildVal.Name,
			expectedName: "childName",
		},
		{
			name:         "Pointer struct field: ChildPtr.Age",
			root:         &p,
			target:       &p.ChildPtr.Age,
			expectedName: "childAge",
		},
		{
			name:         "Anonymous struct field: Anonymous.Inner",
			root:         &p,
			target:       &p.Anonymous.Inner,
			expectedName: "inner",
		},
		// 異常系
		{
			name:          "Root is not a pointer",
			root:          p,
			target:        &p.ID,
			expectedError: true,
		},
		{
			name:          "Root is not a struct pointer",
			root:          &childPtr, // **Child
			target:        &p.ID,
			expectedError: true,
		},
		{
			name:          "Target is not a pointer",
			root:          &p,
			target:        p.ID,
			expectedError: true,
		},
		{
			name:          "Target is unrelated variable",
			root:          &p,
			target:        &otherVar,
			expectedError: true,
		},
		{
			name:         "Target field is nil pointer field itself",
			root:         &p,
			target:       &p.NilPtr,
			expectedName: "nilPtr",
		},
	}

	// intポインタなどをルートにした場合のテスト
	i := 10
	tests = append(tests, struct {
		name          string
		root          any
		target        any
		expectedName  string
		expectedError bool
	}{
		name:          "Root is pointer to int",
		root:          &i,
		target:        &i,
		expectedError: true,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.NameOf(tt.root, tt.target)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if got != tt.expectedName {
					t.Errorf("expected %q, got %q", tt.expectedName, got)
				}
			}
		})
	}
}
