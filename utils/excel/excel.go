package excel

import "io"

// Excel is the common struct holding both Reader and Writer.
// When opened for read (Open/OpenReader), Reader is set and Writer is nil.
// When created for write (NewFile), Writer is set and Reader is nil.
type Excel struct {
	Reader *Reader
	Writer *Writer
}

// Open opens an existing .xlsx file. Use e.Reader to read, then e.Close().
func Open(path string) (*Excel, error) {
	r, err := open(path)
	if err != nil {
		return nil, err
	}
	return &Excel{Reader: r}, nil
}

// OpenReader opens an Excel workbook from rd (e.g. http response body). Use e.Reader to read, then e.Close().
func OpenReader(rd io.Reader) (*Excel, error) {
	r, err := openReader(rd)
	if err != nil {
		return nil, err
	}
	return &Excel{Reader: r}, nil
}

// NewFile creates a new workbook. Use e.Writer to write, then e.Save(path) or e.Close().
func NewFile() *Excel {
	return &Excel{Writer: newFile()}
}

// Close closes the workbook (releases Reader or Writer resources). Safe to call if either is nil.
func (e *Excel) Close() error {
	if e.Reader != nil {
		return e.Reader.Close()
	}
	if e.Writer != nil {
		return e.Writer.Close()
	}
	return nil
}

// Save saves the workbook to path. Only valid when Excel was created with NewFile (Writer is set).
func (e *Excel) Save(path string) error {
	if e.Writer == nil {
		return nil
	}
	return e.Writer.Save(path)
}

// WriteTo writes the workbook to out. Only valid when Writer is set.
func (e *Excel) WriteTo(out io.Writer) (int64, error) {
	if e.Writer == nil {
		return 0, nil
	}
	return e.Writer.WriteTo(out)
}

// ReadFile reads the first sheet of an Excel file and returns all rows.
// For multiple sheets or large files, use Open and e.Reader instead.
func ReadFile(path string) ([][]string, error) {
	e, err := Open(path)
	if err != nil {
		return nil, err
	}
	defer e.Close()
	names := e.Reader.SheetNames()
	if len(names) == 0 {
		return nil, nil
	}
	return e.Reader.ReadSheet(names[0])
}

// ExportFile creates a new workbook, writes rows to "Sheet1", and saves to path.
// For multiple sheets or more control, use NewFile and e.Writer instead.
func ExportFile(path string, rows [][]string) error {
	e := NewFile()
	if err := e.Writer.WriteSheet("Sheet1", rows); err != nil {
		_ = e.Close()
		return err
	}
	err := e.Save(path)
	_ = e.Close()
	return err
}

// ExportTo writes a new workbook with the given rows to out (e.g. http response). Sheet name is "Sheet1".
func ExportTo(out io.Writer, rows [][]string) error {
	e := NewFile()
	if err := e.Writer.WriteSheet("Sheet1", rows); err != nil {
		_ = e.Close()
		return err
	}
	_, err := e.WriteTo(out)
	_ = e.Close()
	return err
}
