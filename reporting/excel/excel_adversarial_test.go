package excel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExcel_GetCellValueByIndex_HighColumns(t *testing.T) {
	path, cleanup := createTempExcelFile(t)
	defer cleanup()

	e := &Excel{}
	require.NoError(t, e.Open(path))
	defer e.Close()

	sheets, err := e.GetSheetList()
	require.NoError(t, err)
	require.NoError(t, e.SetCurrentSheet(sheets[0]))

	// Place distinct values at a double-letter column and the last valid column.
	require.NoError(t, e.book.SetCellValue(sheets[0], "AA1", "double-col"))
	require.NoError(t, e.book.SetCellValue(sheets[0], "XFD1", "last-col"))

	t.Run("double letter column AA (index 26)", func(t *testing.T) {
		got, err := e.GetCellValueByIndex(0, 26)
		require.NoError(t, err)
		assert.Equal(t, "double-col", got)
	})

	t.Run("last column XFD (index 16383)", func(t *testing.T) {
		got, err := e.GetCellValueByIndex(0, 16383)
		require.NoError(t, err)
		assert.Equal(t, "last-col", got)
	})
}

func TestExcel_GetCellValueByIndex_ColumnOutOfRange(t *testing.T) {
	path, cleanup := createTempExcelFile(t)
	defer cleanup()

	e := &Excel{}
	require.NoError(t, e.Open(path))
	defer e.Close()

	sheets, err := e.GetSheetList()
	require.NoError(t, err)
	require.NoError(t, e.SetCurrentSheet(sheets[0]))

	cases := []struct {
		name string
		col  int
	}{
		{"negative index", -1},
		{"MaxColumns is out of range for 0-based index", 16384},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := e.GetCellValueByIndex(0, tc.col)
			assert.Equal(t, "", val)
			assert.ErrorIs(t, err, ColumnIndexOutOfRangeError)
		})
	}
}

func TestExcel_Getters_BookNotOpened(t *testing.T) {
	e := &Excel{}

	t.Run("GetSheetList", func(t *testing.T) {
		list, err := e.GetSheetList()
		assert.Nil(t, list)
		assert.ErrorIs(t, err, BookNotOpenedError)
	})
	t.Run("GetAllCellValues", func(t *testing.T) {
		m, err := e.GetAllCellValues()
		assert.Nil(t, m)
		assert.ErrorIs(t, err, BookNotOpenedError)
	})
	t.Run("GetCellValueByIndex", func(t *testing.T) {
		val, err := e.GetCellValueByIndex(0, 0)
		assert.Equal(t, "", val)
		assert.ErrorIs(t, err, BookNotOpenedError)
	})
	t.Run("GetCellValueByName", func(t *testing.T) {
		val, err := e.GetCellValueByName("A1")
		assert.Equal(t, "", val)
		assert.ErrorIs(t, err, BookNotOpenedError)
	})
	t.Run("Close is no-op", func(t *testing.T) {
		assert.NoError(t, e.Close())
	})
}
