package excel

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Excel struct, Open, OpenReader, NewFile, Close, Save, WriteTo ---

func TestOpen_ReturnsExcelWithReaderSet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "f.xlsx")
	require.NoError(t, ExportFile(path, [][]string{{"A"}}))

	e, err := Open(path)
	require.NoError(t, err)
	defer e.Close()

	assert.NotNil(t, e.Reader)
	assert.Nil(t, e.Writer)
}

func TestOpenReader_ReturnsExcelWithReaderSet(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, ExportTo(&buf, [][]string{{"B"}}))

	e, err := OpenReader(&buf)
	require.NoError(t, err)
	defer e.Close()

	assert.NotNil(t, e.Reader)
	assert.Nil(t, e.Writer)
}

func TestNewFile_ReturnsExcelWithWriterSet(t *testing.T) {
	e := NewFile()
	defer e.Close()

	assert.Nil(t, e.Reader)
	assert.NotNil(t, e.Writer)
}

func TestExcel_Close_WithReader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c.xlsx")
	require.NoError(t, ExportFile(path, [][]string{{"x"}}))
	e, err := Open(path)
	require.NoError(t, err)
	require.NoError(t, e.Close())
	// double close no-op for Reader
	assert.Nil(t, e.Close())
}

func TestExcel_Close_WithWriter(t *testing.T) {
	e := NewFile()
	require.NoError(t, e.Close())
	assert.Nil(t, e.Close())
}

func TestExcel_Save_WithWriter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.xlsx")
	e := NewFile()
	require.NoError(t, e.Writer.WriteSheet("Sheet1", [][]string{{"OK"}}))
	require.NoError(t, e.Save(path))
	e.Close()
	assert.FileExists(t, path)
}

func TestExcel_Save_WithReaderOnly_NoError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "r.xlsx")
	require.NoError(t, ExportFile(path, [][]string{{"x"}}))
	e, err := Open(path)
	require.NoError(t, err)
	defer e.Close()
	// Save when only Reader set does nothing, no error
	assert.NoError(t, e.Save(filepath.Join(t.TempDir(), "out.xlsx")))
}

func TestExcel_WriteTo_WithWriter(t *testing.T) {
	var buf bytes.Buffer
	e := NewFile()
	require.NoError(t, e.Writer.WriteSheet("Sheet1", [][]string{{"V"}}))
	n, err := e.WriteTo(&buf)
	require.NoError(t, err)
	assert.Greater(t, n, int64(0))
	e.Close()
}

func TestExcel_WriteTo_WithReaderOnly_ReturnsZero(t *testing.T) {
	path := filepath.Join(t.TempDir(), "w.xlsx")
	require.NoError(t, ExportFile(path, [][]string{{"x"}}))
	e, err := Open(path)
	require.NoError(t, err)
	defer e.Close()
	var buf bytes.Buffer
	n, err := e.WriteTo(&buf)
	require.NoError(t, err)
	assert.Equal(t, int64(0), n)
}

// --- ReadFile, ExportFile, ExportTo (convenience) ---

func TestReadFile_ReturnsFirstSheetRows(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.xlsx")
	rows := [][]string{
		{"Name", "Age"},
		{"Alice", "30"},
		{"Bob", "25"},
	}
	require.NoError(t, ExportFile(path, rows))

	read, err := ReadFile(path)
	require.NoError(t, err)
	require.Len(t, read, 3)
	assert.Equal(t, rows[0], read[0])
	assert.Equal(t, rows[1], read[1])
	assert.Equal(t, rows[2], read[2])
}

func TestExportFile_CreatesFileWithSheet1(t *testing.T) {
	path := filepath.Join(t.TempDir(), "export.xlsx")
	err := ExportFile(path, [][]string{{"H1", "H2"}, {"a", "b"}})
	require.NoError(t, err)
	assert.FileExists(t, path)

	read, err := ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"H1", "H2"}, {"a", "b"}}, read)
}

func TestExportTo_WritesToWriter(t *testing.T) {
	var buf bytes.Buffer
	rows := [][]string{{"Col1"}, {"Val1"}}
	err := ExportTo(&buf, rows)
	require.NoError(t, err)
	assert.Greater(t, buf.Len(), 0)

	e, err := OpenReader(&buf)
	require.NoError(t, err)
	defer e.Close()
	read, err := e.Reader.ReadSheet("Sheet1")
	require.NoError(t, err)
	assert.Equal(t, rows, read)
}

// --- Error cases ---

func TestReadFile_NotExist(t *testing.T) {
	_, err := ReadFile(filepath.Join(t.TempDir(), "nonexist.xlsx"))
	assert.Error(t, err)
}

func TestOpen_NotExist(t *testing.T) {
	_, err := Open(filepath.Join(t.TempDir(), "nonexist.xlsx"))
	assert.Error(t, err)
}
