package excel

import (
	"fmt"
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

func TestExcel_Open_WithTestData(t *testing.T) {
	path := filepath.Join("..", "..", "_testdata", "reporting", "list.xlsx")

	e := &Excel{}
	if err := e.Open(path); err != nil {
		t.Fatalf("Open failed (path: %s): %v", path, err)
	}
	defer e.Close()

	sheets, err := e.GetSheetList()
	if err != nil {
		t.Errorf("GetSheetList failed: %v", err)
	}
	if len(sheets) == 0 {
		t.Error("expected at least one sheet in testdata")
	}
}

func TestExcel_WriteList_WithTestData(t *testing.T) {
	path := filepath.Join("..", "..", "_testdata", "reporting", "list.xlsx")

	e := &Excel{}
	if err := e.Open(path); err != nil {
		t.Fatalf("Open failed (path: %s): %v", path, err)
	}
	defer e.Close()

	sheetName := "Sheet1"
	if err := e.SetCurrentSheet(sheetName); err != nil {
		t.Fatalf("SetCurrentSheet failed: %v", err)
	}

	// 登録開始行（3行目）
	startRow := 3
	for i := 0; i < 100; i++ {
		row := startRow + i
		no := i + 1
		name := fmt.Sprintf("Item %d", no)
		price := 100 * no
		qty := (i % 10) + 1
		amount := price * qty

		// A列 №
		if err := e.book.SetCellValue(sheetName, fmt.Sprintf("A%d", row), no); err != nil {
			t.Errorf("failed to set No at row %d: %v", row, err)
		}
		// B列 品目名
		if err := e.book.SetCellValue(sheetName, fmt.Sprintf("B%d", row), name); err != nil {
			t.Errorf("failed to set Name at row %d: %v", row, err)
		}
		// C列 単価
		if err := e.book.SetCellValue(sheetName, fmt.Sprintf("C%d", row), price); err != nil {
			t.Errorf("failed to set Price at row %d: %v", row, err)
		}
		// D列 数量
		if err := e.book.SetCellValue(sheetName, fmt.Sprintf("D%d", row), qty); err != nil {
			t.Errorf("failed to set Qty at row %d: %v", row, err)
		}
		// E列 金額
		if err := e.book.SetCellValue(sheetName, fmt.Sprintf("E%d", row), amount); err != nil {
			t.Errorf("failed to set Amount at row %d: %v", row, err)
		}
	}

	savePath := filepath.Join(os.TempDir(), "list_output_test.xlsx")
	if err := e.SaveAs(savePath); err != nil {
		t.Fatalf("SaveAs failed: %v", err)
	}
	defer os.Remove(savePath)

	if _, err := os.Stat(savePath); err != nil {
		t.Errorf("Saved file does not exist: %v", err)
	}

	// 保存したデータを読み込んで検証
	e2 := &Excel{}
	if err := e2.Open(savePath); err != nil {
		t.Fatalf("Open saved file failed: %v", err)
	}
	defer e2.Close()

	if err := e2.SetCurrentSheet(sheetName); err != nil {
		t.Fatalf("SetCurrentSheet (saved file) failed: %v", err)
	}

	for i := 0; i < 100; i++ {
		row := startRow + i
		no := i + 1
		expectedName := fmt.Sprintf("Item %d", no)
		expectedPrice := 100 * no
		expectedQty := (i % 10) + 1
		expectedAmount := expectedPrice * expectedQty

		// A列 №
		valA, err := e2.GetCellValueByName(fmt.Sprintf("A%d", row))
		if err != nil {
			t.Errorf("failed to get A%d: %v", row, err)
		}
		if valA != fmt.Sprintf("%d", no) {
			t.Errorf("row %d col A: expected %d, got %s", row, no, valA)
		}

		// B列 品目名
		valB, err := e2.GetCellValueByName(fmt.Sprintf("B%d", row))
		if err != nil {
			t.Errorf("failed to get B%d: %v", row, err)
		}
		if valB != expectedName {
			t.Errorf("row %d col B: expected %s, got %s", row, expectedName, valB)
		}

		// C列 単価
		valC, err := e2.GetCellValueByName(fmt.Sprintf("C%d", row))
		if err != nil {
			t.Errorf("failed to get C%d: %v", row, err)
		}
		if valC != fmt.Sprintf("%d", expectedPrice) {
			t.Errorf("row %d col C: expected %d, got %s", row, expectedPrice, valC)
		}

		// D列 数量
		valD, err := e2.GetCellValueByName(fmt.Sprintf("D%d", row))
		if err != nil {
			t.Errorf("failed to get D%d: %v", row, err)
		}
		if valD != fmt.Sprintf("%d", expectedQty) {
			t.Errorf("row %d col D: expected %d, got %s", row, expectedQty, valD)
		}

		// E列 金額
		valE, err := e2.GetCellValueByName(fmt.Sprintf("E%d", row))
		if err != nil {
			t.Errorf("failed to get E%d: %v", row, err)
		}
		if valE != fmt.Sprintf("%d", expectedAmount) {
			t.Errorf("row %d col E: expected %d, got %s", row, expectedAmount, valE)
		}
	}
}
