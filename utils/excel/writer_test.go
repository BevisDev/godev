package excel

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper: create new Excel with Writer for tests. Caller must defer e.Close().
func newExcelForWrite(t *testing.T) *Excel {
	t.Helper()
	e := NewFile()
	require.NotNil(t, e.Writer)
	return e
}

func TestWriter_WriteSheet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ws.xlsx")
	e := newExcelForWrite(t)
	defer e.Close()

	err := e.Writer.WriteSheet("Sheet1", [][]string{
		{"H1", "H2", "H3"},
		{"a", "b", "c"},
		{"1", "2", "3"},
	})
	require.NoError(t, err)
	require.NoError(t, e.Save(path))

	read, err := ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"H1", "H2", "H3"}, {"a", "b", "c"}, {"1", "2", "3"}}, read)
}

func TestWriter_WriteSheet_EmptyName_DefaultsToSheet1(t *testing.T) {
	path := filepath.Join(t.TempDir(), "default.xlsx")
	e := newExcelForWrite(t)
	defer e.Close()

	err := e.Writer.WriteSheet("", [][]string{{"X"}})
	require.NoError(t, err)
	require.NoError(t, e.Save(path))

	read, err := ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"X"}}, read)
}

func TestWriter_WriteSheet_CreatesNewSheetIfNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "newsheet.xlsx")
	e := newExcelForWrite(t)
	defer e.Close()

	require.NoError(t, e.Writer.WriteSheet("Report", [][]string{{"Header"}, {"Row1"}}))
	require.NoError(t, e.Save(path))

	e2, err := Open(path)
	require.NoError(t, err)
	defer e2.Close()
	names := e2.Reader.SheetNames()
	assert.Contains(t, names, "Report")
	rows, err := e2.Reader.ReadSheet("Report")
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"Header"}, {"Row1"}}, rows)
}

func TestWriter_SetCell(t *testing.T) {
	path := filepath.Join(t.TempDir(), "setcell.xlsx")
	e := newExcelForWrite(t)
	defer e.Close()

	require.NoError(t, e.Writer.SetCell("Sheet1", "A1", "Title"))
	require.NoError(t, e.Writer.SetCell("Sheet1", "B1", 100))
	require.NoError(t, e.Writer.SetCell("Sheet1", "A2", true))
	require.NoError(t, e.Save(path))

	e2, err := Open(path)
	require.NoError(t, err)
	defer e2.Close()
	v, err := e2.Reader.GetCell("Sheet1", "A1")
	require.NoError(t, err)
	assert.Equal(t, "Title", v)
	v, err = e2.Reader.GetCell("Sheet1", "B1")
	require.NoError(t, err)
	assert.Equal(t, "100", v)
	v, err = e2.Reader.GetCell("Sheet1", "A2")
	require.NoError(t, err)
	assert.Equal(t, "TRUE", v)
}

func TestWriter_AddSheet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "add.xlsx")
	e := newExcelForWrite(t)
	defer e.Close()

	idx, err := e.Writer.AddSheet("Extra")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, idx, 0)

	require.NoError(t, e.Writer.WriteSheet("Extra", [][]string{{"in"}, {"extra"}}))
	require.NoError(t, e.Save(path))

	e2, err := Open(path)
	require.NoError(t, err)
	defer e2.Close()
	assert.Contains(t, e2.Reader.SheetNames(), "Extra")
	rows, err := e2.Reader.ReadSheet("Extra")
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"in"}, {"extra"}}, rows)
}

func TestWriter_SetActiveSheet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "active.xlsx")
	e := newExcelForWrite(t)
	defer e.Close()

	require.NoError(t, e.Writer.WriteSheet("Sheet1", [][]string{{"First"}}))
	idx, err := e.Writer.AddSheet("Second")
	require.NoError(t, err)
	require.NoError(t, e.Writer.WriteSheet("Second", [][]string{{"Second"}}))
	e.Writer.SetActiveSheet(idx)
	require.NoError(t, e.Save(path))
	e.Close()
	// File saved; active sheet is Second (index 1). Opening and reading still works.
	e2, err := Open(path)
	require.NoError(t, err)
	defer e2.Close()
	rows, err := e2.Reader.ReadSheet("Second")
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"Second"}}, rows)
}

func TestWriter_Save(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.xlsx")
	e := newExcelForWrite(t)
	require.NoError(t, e.Writer.WriteSheet("Sheet1", [][]string{{"OK"}}))
	err := e.Writer.Save(path)
	require.NoError(t, err)
	e.Close()
	assert.FileExists(t, path)
}

func TestWriter_WriteTo(t *testing.T) {
	e := newExcelForWrite(t)
	defer e.Close()
	require.NoError(t, e.Writer.WriteSheet("Sheet1", [][]string{{"Stream"}}))

	var buf bytes.Buffer
	n, err := e.Writer.WriteTo(&buf)
	require.NoError(t, err)
	assert.Greater(t, n, int64(0))
	assert.Greater(t, buf.Len(), 0)

	e2, err := OpenReader(&buf)
	require.NoError(t, err)
	defer e2.Close()
	rows, err := e2.Reader.ReadSheet("Sheet1")
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"Stream"}}, rows)
}

func TestWriter_Close(t *testing.T) {
	e := newExcelForWrite(t)
	err := e.Writer.Close()
	require.NoError(t, err)
	e.Writer = nil
	assert.Nil(t, e.Close())
}
