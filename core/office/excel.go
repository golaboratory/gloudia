package office

import (
	"errors"
	"github.com/xuri/excelize/v2"
	"golang.org/x/exp/slices"
	"strconv"
)

var (
	BookNotOpenedError         = errors.New("book is not opened")
	SheetNotFoundError         = errors.New("sheet not found")
	ColumnIndexOutOfRangeError = errors.New("column index out of range")
	columnsLetters             = map[int]string{
		0:  "A",
		1:  "B",
		2:  "C",
		3:  "D",
		4:  "E",
		5:  "F",
		6:  "G",
		7:  "H",
		8:  "I",
		9:  "J",
		10: "K",
		11: "L",
		12: "M",
		13: "N",
		14: "O",
		15: "P",
		16: "Q",
		17: "R",
		18: "S",
		19: "T",
		20: "U",
		21: "V",
		22: "W",
		23: "X",
		24: "Y",
		25: "Z",
	}
)

// CellPosition はセルの位置（行・列）を表します。
type CellPosition struct {
	Row    int // 行番号（0始まり）
	Column int // 列番号（0始まり）
}

// Excel は Excel ファイルを操作するための構造体です。
// ファイルのオープン、シート一覧取得、シートコピー、保存、クローズなどの機能を提供します。
type Excel struct {
	book             *excelize.File // Excelファイルのハンドル
	FilePath         string         // ファイルパス
	CurrentSheetName string         // 現在のシート名
}

// Open は指定されたパスの Excel ファイルを開きます。
// ファイルのオープンに失敗した場合はエラーを返します。
func (e *Excel) Open(path string) error {
	var err error
	e.book, err = excelize.OpenFile(path)
	if err != nil {
		return err
	}
	e.FilePath = path // ファイルパスを設定
	return nil
}

// GetSheetList は開かれている Excel ファイルのシート名一覧を取得します。
// ファイルが開かれていない場合はエラーを返します。
func (e *Excel) GetSheetList() ([]string, error) {
	if e.book == nil {
		return nil, BookNotOpenedError
	}
	return e.book.GetSheetList(), nil
}

// CopySheet は src で指定したシートを dest という名前でコピーします。
// コピー元シートが存在しない場合やコピー先シート名が既に存在する場合はエラーを返します。
func (e *Excel) CopySheet(src, dest string) error {
	if e.book == nil {
		return BookNotOpenedError
	}

	if !e.existSheet(src) {
		return SheetNotFoundError
	}
	if e.existSheet(dest) {
		return SheetNotFoundError
	}

	srcIndex, err := e.book.GetSheetIndex(src)
	if err != nil {
		return err
	}

	destIndex, err := e.book.NewSheet(dest)
	if err != nil {
		return err
	}

	return e.book.CopySheet(srcIndex, destIndex)
}

// SetCurrentSheet は指定したシート名を現在のシートとして設定します。
// シートが存在しない場合はエラーを返します。
func (e *Excel) SetCurrentSheet(sheetName string) error {
	if e.book == nil {
		return BookNotOpenedError
	}

	if !e.existSheet(sheetName) {
		return SheetNotFoundError
	}

	e.CurrentSheetName = sheetName
	return nil
}

// GetAllCellValues は現在のシートの全セルの値を取得します。
// セル位置（行・列）と値のマップを返します。
// シートが存在しない場合はエラーを返します。
func (e *Excel) GetAllCellValues() (map[CellPosition]string, error) {
	if e.book == nil {
		return nil, BookNotOpenedError
	}

	if !e.existSheet(e.CurrentSheetName) {
		return nil, SheetNotFoundError
	}

	cells := make(map[CellPosition]string)

	rows, err := e.book.GetRows(e.CurrentSheetName)
	if err != nil {
		return nil, err
	}

	for rowIndex, row := range rows {
		for columnIndex, colCell := range row {
			cells[CellPosition{Row: rowIndex, Column: columnIndex}] = colCell
		}
	}

	return cells, nil
}

// GetCellValueByIndex は指定した行・列インデックスのセル値を取得します。
// 行・列は0始まりです。シートが存在しない場合はエラーを返します。
func (e *Excel) GetCellValueByIndex(rowIndex, columnIndex int) (string, error) {
	if e.book == nil {
		return "", BookNotOpenedError
	}

	if !e.existSheet(e.CurrentSheetName) {
		return "", SheetNotFoundError
	}

	colName, err := convertColumnIndexToLetter(columnIndex)
	if err != nil {
		return "", err
	}

	return e.GetCellValueByName(colName + strconv.Itoa(rowIndex+1))
}

// GetCellValueByName はセル名（例: "A1"）で指定したセルの値を取得します。
// シートが存在しない場合はエラーを返します。
func (e *Excel) GetCellValueByName(cellName string) (string, error) {
	if e.book == nil {
		return "", BookNotOpenedError
	}

	if !e.existSheet(e.CurrentSheetName) {
		return "", SheetNotFoundError
	}

	value, err := e.book.GetCellValue(e.CurrentSheetName, cellName)
	if err != nil {
		return "", err
	}
	return value, nil

}

// SaveAs は Excel ファイルを指定されたパスに保存します。
// ファイルが開かれていない場合はエラーを返します。
func (e *Excel) SaveAs(path string) error {
	if e.book == nil {
		return BookNotOpenedError
	}
	return e.book.SaveAs(path)
}

// Save は現在開いている Excel ファイルを上書き保存します。
// ファイルが開かれていない場合はエラーを返します。
func (e *Excel) Save() error {
	if e.book == nil {
		return BookNotOpenedError
	}
	return e.book.Save()
}

// Close は開いている Excel ファイルをクローズします。
// ファイルが開かれていない場合は何もしません。
func (e *Excel) Close() error {
	if e.book == nil {
		return nil
	}
	return e.book.Close()
}

// existSheet は指定したシート名が存在するかどうかを判定します。
// 存在する場合は true、存在しない場合は false を返します。
func (e *Excel) existSheet(sheetName string) bool {
	if e.book == nil {
		return false
	}
	sheetNames, err := e.GetSheetList()
	if err != nil {
		return false
	}
	return slices.Contains(sheetNames, sheetName)
}

// convertColumnIndexToLetter は列インデックス（0始まり）をExcelの列名（例: "A", "AB"）に変換します。
// 範囲外の場合はエラーを返します。
func convertColumnIndexToLetter(index int) (string, error) {
	if index < 0 || index > excelize.MaxColumns {
		return "", ColumnIndexOutOfRangeError
	}

	if index < 26 {
		return columnsLetters[index], nil
	}

	if index < (26*26 + 26) {
		return columnsLetters[index/26-1] + columnsLetters[index%26], nil
	}

	if index < excelize.MaxColumns {
		return columnsLetters[index/26/26-1] + columnsLetters[(index/26)%26] + columnsLetters[index%26], nil
	}
	return "", ColumnIndexOutOfRangeError
}
