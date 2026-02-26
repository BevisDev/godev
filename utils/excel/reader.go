package excel

import (
	"io"

	"github.com/xuri/excelize/v2"
)

// Reader wraps an open Excel file for reading.
type Reader struct {
	f *excelize.File
}

// open opens an existing .xlsx file for reading. Used by Excel.Open.
func open(path string) (*Reader, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	return &Reader{f: f}, nil
}

// openReader opens an Excel workbook from r (e.g. http response body). Used by Excel.OpenReader.
func openReader(r io.Reader) (*Reader, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}
	return &Reader{f: f}, nil
}

// SheetNames returns the names of all sheets in the workbook.
func (r *Reader) SheetNames() []string {
	return r.f.GetSheetList()
}

// ReadSheet returns all rows in the given sheet. Each row is a []string of cell values.
// Empty trailing cells may be omitted (row length can vary).
func (r *Reader) ReadSheet(sheetName string) ([][]string, error) {
	return r.f.GetRows(sheetName)
}

// ReadSheetAt returns all rows of the sheet at index (0-based). Same as ReadSheet(names[index]).
func (r *Reader) ReadSheetAt(index int) ([][]string, error) {
	names := r.SheetNames()
	if index < 0 || index >= len(names) {
		return nil, nil
	}
	return r.ReadSheet(names[index])
}

// GetCell returns the value of a single cell (e.g. "A1", "B2").
func (r *Reader) GetCell(sheetName, cell string) (string, error) {
	return r.f.GetCellValue(sheetName, cell)
}

// Close closes the workbook and releases resources.
func (r *Reader) Close() error {
	return r.f.Close()
}
