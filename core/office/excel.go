package office

import (
	"errors"
	"github.com/xuri/excelize/v2"
	"golang.org/x/exp/slices"
)

// Excel は Excel ファイルを操作するための構造体です。
// ファイルのオープン、シート一覧取得、シートコピー、保存、クローズなどの機能を提供します。
type Excel struct {
	Book *excelize.File
}

// Open は指定されたパスの Excel ファイルを開きます。
// ファイルのオープンに失敗した場合はエラーを返します。
func (e *Excel) Open(path string) error {
	var err error
	e.Book, err = excelize.OpenFile(path)
	if err != nil {
		return err
	}
	return nil
}

// GetSheetList は開かれている Excel ファイルのシート名一覧を取得します。
// ファイルが開かれていない場合はエラーを返します。
func (e *Excel) GetSheetList() ([]string, error) {
	if e.Book == nil {
		return nil, errors.New("book is not opened")
	}
	return e.Book.GetSheetList(), nil
}

// CopySheet は src で指定したシートを dest という名前でコピーします。
// コピー元シートが存在しない場合やコピー先シート名が既に存在する場合はエラーを返します。
func (e *Excel) CopySheet(src, dest string) error {
	if e.Book == nil {
		return errors.New("book is not opened")
	}

	sheetNames, err := e.GetSheetList()
	if err != nil {
		return err
	}

	if !slices.Contains(sheetNames, src) {
		return errors.New("source sheet does not exist")
	}

	if slices.Contains(sheetNames, dest) {
		return errors.New("destination sheet already exists")
	}

	srcIndex, err := e.Book.GetSheetIndex(src)
	if err != nil {
		return err
	}

	destIndex, err := e.Book.NewSheet(dest)
	if err != nil {
		return err
	}

	return e.Book.CopySheet(srcIndex, destIndex)
}

// SaveAs は Excel ファイルを指定されたパスに保存します。
// ファイルが開かれていない場合はエラーを返します。
func (e *Excel) SaveAs(path string) error {
	if e.Book == nil {
		return errors.New("book is not opened")
	}
	return e.Book.SaveAs(path)
}

// Save は現在開いている Excel ファイルを上書き保存します。
// ファイルが開かれていない場合はエラーを返します。
func (e *Excel) Save() error {
	if e.Book == nil {
		return errors.New("book is not opened")
	}
	return e.Book.Save()
}

// Close は開いている Excel ファイルをクローズします。
// ファイルが開かれていない場合は何もしません。
func (e *Excel) Close() error {
	if e.Book == nil {
		return nil
	}
	return e.Book.Close()
}
