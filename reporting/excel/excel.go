package excel

import (
	"slices"
	"strconv"

	"github.com/newmo-oss/ergo"
	"github.com/xuri/excelize/v2"
)

// センチネルエラー定義
var (
	// ErrBookNotOpened は Open 前にファイル操作を行った場合のエラーです。
	ErrBookNotOpened = ergo.NewSentinel("book is not opened")
	// ErrSheetNotFound は指定した名前のシートが存在しない場合のエラーです。
	ErrSheetNotFound = ergo.NewSentinel("sheet not found")
	// ErrSheetAlreadyExists はコピー先など、既に存在するシート名を指定した場合のエラーです。
	ErrSheetAlreadyExists = ergo.NewSentinel("sheet already exists")
	// ErrColumnIndexOutOfRange は列番号が Excel の上限 (XFD = 16383) を超えた場合や
	// 負の場合のエラーです。
	ErrColumnIndexOutOfRange = ergo.NewSentinel("column index out of range")
)

// 以下は Go の慣習 (Err プレフィックス) に合わせる前の旧名エイリアスです。
// 同一のセンチネル値のため errors.Is での判定結果は変わりません。
var (
	// Deprecated: ErrBookNotOpened を使用してください。
	BookNotOpenedError = ErrBookNotOpened
	// Deprecated: ErrSheetNotFound を使用してください。
	SheetNotFoundError = ErrSheetNotFound
	// Deprecated: ErrColumnIndexOutOfRange を使用してください。
	ColumnIndexOutOfRangeError = ErrColumnIndexOutOfRange
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
		return nil, ErrBookNotOpened
	}
	return e.book.GetSheetList(), nil
}

// CopySheet は src で指定したシートを dest という名前でコピーします。
// コピー元シートが存在しない場合は ErrSheetNotFound を、
// コピー先シート名が既に存在する場合は ErrSheetAlreadyExists を返します。
func (e *Excel) CopySheet(src, dest string) error {
	if e.book == nil {
		return ErrBookNotOpened
	}

	if !e.existSheet(src) {
		return ErrSheetNotFound
	}
	if e.existSheet(dest) {
		return ErrSheetAlreadyExists
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
		return ErrBookNotOpened
	}

	if !e.existSheet(sheetName) {
		return ErrSheetNotFound
	}

	e.CurrentSheetName = sheetName
	return nil
}

// GetAllCellValues は現在のシートの全セルの値を取得します。
// セル位置（行・列）と値のマップを返します。
// シートが存在しない場合はエラーを返します。
func (e *Excel) GetAllCellValues() (map[CellPosition]string, error) {
	if e.book == nil {
		return nil, ErrBookNotOpened
	}

	if !e.existSheet(e.CurrentSheetName) {
		return nil, ErrSheetNotFound
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
		return "", ErrBookNotOpened
	}

	if !e.existSheet(e.CurrentSheetName) {
		return "", ErrSheetNotFound
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
		return "", ErrBookNotOpened
	}

	if !e.existSheet(e.CurrentSheetName) {
		return "", ErrSheetNotFound
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
		return ErrBookNotOpened
	}
	return e.book.SaveAs(path)
}

// Save は現在開いている Excel ファイルを上書き保存します。
// ファイルが開かれていない場合はエラーを返します。
func (e *Excel) Save() error {
	if e.book == nil {
		return ErrBookNotOpened
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
//
// 以前の手書き実装は 3 文字目（index>=702, "AAA" 以降）で桁上がりを誤り、
// 全列の大半で誤った列名を返していた。bijective base-26 を正しく実装している
// excelize.ColumnNumberToName（1始まり）に委譲する。
func convertColumnIndexToLetter(index int) (string, error) {
	// 有効な 0 始まりインデックスは 0..MaxColumns-1。MaxColumns(16384) は範囲外。
	if index < 0 || index >= excelize.MaxColumns {
		return "", ErrColumnIndexOutOfRange
	}
	name, err := excelize.ColumnNumberToName(index + 1)
	if err != nil {
		return "", ErrColumnIndexOutOfRange
	}
	return name, nil
}
