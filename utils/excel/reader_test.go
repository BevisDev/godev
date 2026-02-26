package excel

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper: open a file and return e.Reader (caller must defer e.Close()).
func openExcelForRead(t *testing.T, path string) (*Excel, *Reader) {
	t.Helper()
	e, err := Open(path)
	require.NoError(t, err)
	require.NotNil(t, e.Reader)
	return e, e.Reader
}

func TestReader_SheetNames(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sheets.xlsx")
	e := NewFile()
	require.NoError(t, e.Writer.WriteSheet("Sheet1", [][]string{{"a"}}))
	_, err := e.Writer.AddSheet("Data")
	require.NoError(t, err)
	require.NoError(t, e.Writer.WriteSheet("Data", [][]string{{"x"}}))
	require.NoError(t, e.Save(path))
	require.NoError(t, e.Close())

	e2, err := Open(path)
	require.NoError(t, err)
	defer e2.Close()
	r := e2.Reader

	names := r.SheetNames()
	assert.Contains(t, names, "Sheet1")
	assert.Contains(t, names, "Data")
	assert.Len(t, names, 2)
}

func TestReader_ReadSheet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "read.xlsx")
	require.NoError(t, ExportFile(path, [][]string{{"A", "B"}, {"1", "2"}}))

	e, r := openExcelForRead(t, path)
	defer e.Close()

	rows, err := r.ReadSheet("Sheet1")
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, []string{"A", "B"}, rows[0])
	assert.Equal(t, []string{"1", "2"}, rows[1])
}

func TestReader_ReadSheet_MultipleSheets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "multi.xlsx")
	e := NewFile()
	require.NoError(t, e.Writer.WriteSheet("First", [][]string{{"only"}}))
	_, err := e.Writer.AddSheet("Second")
	require.NoError(t, err)
	require.NoError(t, e.Writer.WriteSheet("Second", [][]string{{"a", "b"}}))
	require.NoError(t, e.Save(path))
	e.Close()

	e2, r := openExcelForRead(t, path)
	defer e2.Close()

	first, err := r.ReadSheet("First")
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"only"}}, first)

	second, err := r.ReadSheet("Second")
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"a", "b"}}, second)
}

func TestReader_ReadSheetAt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "at.xlsx")
	require.NoError(t, ExportFile(path, [][]string{{"Row0"}}))

	e, r := openExcelForRead(t, path)
	defer e.Close()

	rows0, err := r.ReadSheetAt(0)
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"Row0"}}, rows0)
}

func TestReader_ReadSheetAt_OutOfRange_ReturnsNil(t *testing.T) {
	path := filepath.Join(t.TempDir(), "one.xlsx")
	require.NoError(t, ExportFile(path, [][]string{{"Only"}}))

	e, r := openExcelForRead(t, path)
	defer e.Close()

	rows, err := r.ReadSheetAt(10)
	require.NoError(t, err)
	assert.Nil(t, rows)
}

func TestReader_ReadSheetAt_NegativeIndex_ReturnsNil(t *testing.T) {
	path := filepath.Join(t.TempDir(), "neg.xlsx")
	require.NoError(t, ExportFile(path, [][]string{{"X"}}))

	e, r := openExcelForRead(t, path)
	defer e.Close()

	rows, err := r.ReadSheetAt(-1)
	require.NoError(t, err)
	assert.Nil(t, rows)
}

func TestReader_GetCell(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cell.xlsx")
	require.NoError(t, ExportFile(path, [][]string{{"Hello", "World"}, {"A", "B"}}))

	e, r := openExcelForRead(t, path)
	defer e.Close()

	val, err := r.GetCell("Sheet1", "A1")
	require.NoError(t, err)
	assert.Equal(t, "Hello", val)

	val, err = r.GetCell("Sheet1", "B1")
	require.NoError(t, err)
	assert.Equal(t, "World", val)

	val, err = r.GetCell("Sheet1", "A2")
	require.NoError(t, err)
	assert.Equal(t, "A", val)
}

func TestReader_OpenFromReader(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, ExportTo(&buf, [][]string{{"From"}, {"Buffer"}}))

	e, err := OpenReader(&buf)
	require.NoError(t, err)
	defer e.Close()
	r := e.Reader

	names := r.SheetNames()
	require.Len(t, names, 1)
	rows, err := r.ReadSheet(names[0])
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"From"}, {"Buffer"}}, rows)
}

func TestReader_Close(t *testing.T) {
	path := filepath.Join(t.TempDir(), "close.xlsx")
	require.NoError(t, ExportFile(path, [][]string{{"x"}}))
	e, r := openExcelForRead(t, path)
	require.NoError(t, r.Close())
	// e.Close() would double-close; e.Reader is still non-nil
	e.Reader = nil
	assert.Nil(t, e.Close())
}
