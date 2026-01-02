package excel

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
	_, err := f.NewSheet(sheetName)
	if err != nil {
		t.Fatalf("failed to create new sheet: %v", err)
	}
	path := filepath.Join(os.TempDir(), "test_excel.xlsx")
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("failed to save temp excel file: %v", err)
	}
	return path, func() {
		err := os.Remove(path)
		if err != nil && !os.IsNotExist(err) {
			t.Errorf("failed to remove temp excel file: %v", err)
		}
	}
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

func TestExcel_SetCurrentSheet(t *testing.T) {
	path, cleanup := createTempExcelFile(t)
	defer cleanup()

	e := &Excel{}
	if err := e.Open(path); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	sheets, _ := e.GetSheetList()
	sheet := sheets[0]
	if err := e.SetCurrentSheet(sheet); err != nil {
		t.Errorf("SetCurrentSheet failed: %v", err)
	}
	if e.CurrentSheetName != sheet {
		t.Errorf("CurrentSheetName not set correctly")
	}
	// 存在しないシート名
	if err := e.SetCurrentSheet("NoSheet"); err == nil {
		t.Error("SetCurrentSheet should fail for non-existent sheet")
	}
}

func TestExcel_GetAllCellValues(t *testing.T) {
	path, cleanup := createTempExcelFile(t)
	defer cleanup()

	e := &Excel{}
	if err := e.Open(path); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	sheets, _ := e.GetSheetList()
	sheet := sheets[0]
	err := e.SetCurrentSheet(sheet)
	if err != nil {
		t.Fatalf("SetCurrentSheet failed: %v", err)
	}
	// セルに値をセット
	_ = e.book.SetCellValue(sheet, "A1", "foo")
	_ = e.book.SetCellValue(sheet, "B2", "bar")
	values, err := e.GetAllCellValues()
	if err != nil {
		t.Errorf("GetAllCellValues failed: %v", err)
	}
	foundA1, foundB2 := false, false
	for pos, val := range values {
		if pos.Row == 0 && pos.Column == 0 && val == "foo" {
			foundA1 = true
		}
		if pos.Row == 1 && pos.Column == 1 && val == "bar" {
			foundB2 = true
		}
	}
	if !foundA1 || !foundB2 {
		t.Error("Cell values not found as expected")
	}
}

func TestExcel_GetCellValueByIndex(t *testing.T) {
	path, cleanup := createTempExcelFile(t)
	defer cleanup()

	e := &Excel{}
	if err := e.Open(path); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	sheets, _ := e.GetSheetList()
	sheet := sheets[0]
	err := e.SetCurrentSheet(sheet)
	if err != nil {
		t.Fatalf("SetCurrentSheet failed: %v", err)
	}
	_ = e.book.SetCellValue(sheet, "C3", "baz")
	val, err := e.GetCellValueByIndex(2, 2) // C3
	if err != nil {
		t.Errorf("GetCellValueByIndex failed: %v", err)
	}
	if val != "baz" {
		t.Errorf("expected 'baz', got '%s'", val)
	}
}

func TestExcel_GetCellValueByName(t *testing.T) {
	path, cleanup := createTempExcelFile(t)
	defer cleanup()

	e := &Excel{}
	if err := e.Open(path); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	sheets, _ := e.GetSheetList()
	sheet := sheets[0]
	err := e.SetCurrentSheet(sheet)
	if err != nil {
		t.Fatalf("SetCurrentSheet failed: %v", err)
	}
	_ = e.book.SetCellValue(sheet, "D4", "qux")
	val, err := e.GetCellValueByName("D4")
	if err != nil {
		t.Errorf("GetCellValueByName failed: %v", err)
	}
	if val != "qux" {
		t.Errorf("expected 'qux', got '%s'", val)
	}
}

func TestExcel_existSheet(t *testing.T) {
	path, cleanup := createTempExcelFile(t)
	defer cleanup()

	e := &Excel{}
	if err := e.Open(path); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	sheets, _ := e.GetSheetList()
	if !e.existSheet(sheets[0]) {
		t.Error("existSheet should return true for existing sheet")
	}
	if e.existSheet("NoSheet") {
		t.Error("existSheet should return false for non-existent sheet")
	}
}

func Test_convertColumnIndexToLetter(t *testing.T) {
	tests := []struct {
		index int
		want  string
	}{
		{0, "A"},
		{25, "Z"},
		{26, "AA"},
		{27, "AB"},
		{51, "AZ"},
		{52, "BA"},
		{701, "ZZ"},
	}
	for _, tt := range tests {
		got, err := convertColumnIndexToLetter(tt.index)
		if err != nil {
			t.Errorf("convertColumnIndexToLetter(%d) error: %v", tt.index, err)
		}
		if got != tt.want {
			t.Errorf("convertColumnIndexToLetter(%d) = %s, want %s", tt.index, got, tt.want)
		}
	}
	// 範囲外
	if _, err := convertColumnIndexToLetter(-1); err == nil {
		t.Error("convertColumnIndexToLetter should fail for negative index")
	}
}
