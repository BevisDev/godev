package excel

import (
	"io"

	"github.com/xuri/excelize/v2"
)

// Writer wraps a new workbook for exporting.
type Writer struct {
	f *excelize.File
}

// newFile creates a new workbook. Used by Excel.NewFile.
func newFile() *Writer {
	return &Writer{f: excelize.NewFile()}
}

// WriteSheet writes rows to the given sheet. If the sheet does not exist, it is created.
// rows[i] is row i+1; rows[i][j] is column j+1. Supports string, number, bool, empty string.
func (w *Writer) WriteSheet(sheetName string, rows [][]string) error {
	if sheetName == "" {
		sheetName = "Sheet1"
	}
	idx, err := w.f.GetSheetIndex(sheetName)
	if err != nil || idx < 0 {
		_, err = w.f.NewSheet(sheetName)
		if err != nil {
			return err
		}
	}
	for rowIdx, row := range rows {
		for colIdx, val := range row {
			cell, err := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
			if err != nil {
				return err
			}
			if err := w.f.SetCellValue(sheetName, cell, val); err != nil {
				return err
			}
		}
	}
	return nil
}

// SetCell sets the value of a single cell (e.g. "A1", "B2"). Value can be string, int, float64, bool, etc.
func (w *Writer) SetCell(sheetName, cell string, value interface{}) error {
	return w.f.SetCellValue(sheetName, cell, value)
}

// AddSheet adds a new sheet with the given name. Use WriteSheet or SetCell to fill it.
func (w *Writer) AddSheet(name string) (int, error) {
	return w.f.NewSheet(name)
}

// SetActiveSheet sets the active (default) sheet when the file is opened (0-based index).
func (w *Writer) SetActiveSheet(index int) {
	w.f.SetActiveSheet(index)
}

// Save writes the workbook to path (e.g. "output.xlsx").
func (w *Writer) Save(path string) error {
	return w.f.SaveAs(path)
}

// WriteTo writes the workbook to out. Caller may use this to stream to http.ResponseWriter.
func (w *Writer) WriteTo(out io.Writer) (int64, error) {
	return w.f.WriteTo(out)
}

// Close releases resources. Call when done if you do not call Save/WriteTo.
func (w *Writer) Close() error {
	return w.f.Close()
}
