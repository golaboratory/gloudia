package office

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

func createTempExcelFile(t *testing.T) (string, func()) {
	t.Helper()
	f := excelize.NewFile()
	sheetName := "Sheet1"
	f.NewSheet(sheetName)
	path := filepath.Join(os.TempDir(), "test_excel.xlsx")
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("failed to save temp excel file: %v", err)
	}
	return path, func() { os.Remove(path) }
}

func TestExcel_Open(t *testing.T) {
	path, cleanup := createTempExcelFile(t)
	defer cleanup()

	e := &Excel{}
	if err := e.Open(path); err != nil {
		t.Errorf("Open failed: %v", err)
	}
}

func TestExcel_GetSheetList(t *testing.T) {
	path, cleanup := createTempExcelFile(t)
	defer cleanup()

	e := &Excel{}
	if err := e.Open(path); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	sheets, err := e.GetSheetList()
	if err != nil {
		t.Errorf("GetSheetList failed: %v", err)
	}
	if len(sheets) == 0 {
		t.Error("expected at least one sheet")
	}
}

func TestExcel_CopySheet(t *testing.T) {
	path, cleanup := createTempExcelFile(t)
	defer cleanup()

	e := &Excel{}
	if err := e.Open(path); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	sheets, _ := e.GetSheetList()
	src := sheets[0]
	dest := "CopiedSheet"
	if err := e.CopySheet(src, dest); err != nil {
		t.Errorf("CopySheet failed: %v", err)
	}
	// コピー先が存在するか確認
	sheets, _ = e.GetSheetList()
	found := false
	for _, s := range sheets {
		if s == dest {
			found = true
			break
		}
	}
	if !found {
		t.Error("Copied sheet not found")
	}
}

func TestExcel_SaveAs(t *testing.T) {
	path, cleanup := createTempExcelFile(t)
	defer cleanup()

	e := &Excel{}
	if err := e.Open(path); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	savePath := filepath.Join(os.TempDir(), "test_excel_saveas.xlsx")
	defer os.Remove(savePath)
	if err := e.SaveAs(savePath); err != nil {
		t.Errorf("SaveAs failed: %v", err)
	}
	if _, err := os.Stat(savePath); err != nil {
		t.Errorf("SaveAs did not create file: %v", err)
	}
}

func TestExcel_Save(t *testing.T) {
	path, cleanup := createTempExcelFile(t)
	defer cleanup()

	e := &Excel{}
	if err := e.Open(path); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if err := e.Save(); err != nil {
		t.Errorf("Save failed: %v", err)
	}
}

func TestExcel_Close(t *testing.T) {
	path, cleanup := createTempExcelFile(t)
	defer cleanup()

	e := &Excel{}
	if err := e.Open(path); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if err := e.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
	// 2回目のCloseは何もしないことを確認
	if err := e.Close(); err != nil {
		t.Errorf("Close (second) failed: %v", err)
	}
}
