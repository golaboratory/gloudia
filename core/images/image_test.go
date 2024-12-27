package images

import (
	"fmt"
	"os"
	"runtime"
	"testing"
)

// TestResizeToData は、ResizeToData メソッドのテストを行います。
func TestResizeToData(t *testing.T) {
	tests := []struct {
		filePath string
		fileType string
	}{
		{"../../testdata/core/images/test.png", "png"},
		{"../../testdata/core/images/test.jpg", "jpeg"},
	}
	_, testSourceFilePath, _, _ := runtime.Caller(0)
	fmt.Println(testSourceFilePath)

	for _, tt := range tests {
		t.Run(tt.fileType, func(t *testing.T) {
			img := Image{
				FilePath:    tt.filePath,
				ChangeRatio: 0.5,
				Target:      RATIO,
			}

			data, err := img.ResizeToData()
			if err != nil {
				t.Fatalf("ResizeToData returned an error: %v", err)
			}

			if len(data) == 0 {
				t.Fatalf("ResizeToData returned empty data")
			}
		})
	}
}

// TestResizeToFile は、ResizeToFile メソッドのテストを行います。
func TestResizeToFile(t *testing.T) {
	tests := []struct {
		filePath string
		fileType string
	}{
		{"../../testdata/core/images/test.png", "png"},
		{"../../testdata/core/images/test.jpg", "jpeg"},
	}

	for _, tt := range tests {
		t.Run(tt.fileType, func(t *testing.T) {
			img := Image{
				FilePath:    tt.filePath,
				ChangeWidth: 100,
				Target:      WIDTH,
			}

			filePath, err := img.ResizeToFile()
			if err != nil {
				t.Fatalf("ResizeToFile returned an error: %v", err)
			}

			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Fatalf("ResizeToFile did not create the file")
			}

			fmt.Println(filePath)

			// Clean up
			//defer os.Remove(filePath)
		})
	}
}
