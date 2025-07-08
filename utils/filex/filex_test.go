package filex

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	CreateDirName = "createdir"
	PrefixTempDir = "tempdir_"
)

func createTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return path
}

func TestReadAsBytesAndString(t *testing.T) {
	dir, err := CreateDirTemp(PrefixTempDir)
	content := "hello world"
	filePath := createTempFile(t, dir, "file.txt", content)

	b, err := ReadAsBytes(filePath)
	if err != nil {
		t.Fatalf("ReadAsBytes error: %v", err)
	}
	if string(b) != content {
		t.Errorf("ReadAsBytes got %q, want %q", string(b), content)
	}

	s, err := ReadAsString(filePath)
	if err != nil {
		t.Fatalf("ReadAsString error: %v", err)
	}
	if s != content {
		t.Errorf("ReadAsString got %q, want %q", s, content)
	}
}

func TestReadAsLines(t *testing.T) {
	dir := t.TempDir()
	content := "line1\nline2\nline3\n"
	filePath := createTempFile(t, dir, "file.txt", content)

	lines, err := ReadAsLines(filePath)
	if err != nil {
		t.Fatalf("ReadAsLines error: %v", err)
	}

	want := []string{"line1", "line2", "line3"}
	if len(lines) != len(want) {
		t.Fatalf("ReadAsLines got %d lines, want %d", len(lines), len(want))
	}
	for i := range lines {
		if lines[i] != want[i] {
			t.Errorf("line %d got %q, want %q", i, lines[i], want[i])
		}
	}
}

func TestCreateDir(t *testing.T) {
	dir := filepath.Join(os.TempDir(), CreateDirName)
	defer DelAll(dir)

	err := CreateDir(dir)
	if err != nil {
		t.Fatalf("CreateDir failed: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Failed to stat created dir: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("Created path is not a directory")
	}
}

func TestCreateDirTemp(t *testing.T) {
	dir, err := CreateDirTemp(PrefixTempDir)
	if err != nil {
		t.Fatalf("CreateDirTemp failed: %v", err)
	}
	defer DelAll(dir)

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Failed to stat temp dir: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("Created temp path is not a directory")
	}
}

func TestSaveFileAndIsDir(t *testing.T) {
	dir := t.TempDir()
	content := []byte("test content")
	filename := "testfile.txt"

	err := SaveFile(dir, filename, content)
	if err != nil {
		t.Fatalf("SaveFile error: %v", err)
	}

	path := filepath.Join(dir, filename)

	b, err := ReadAsBytes(path)
	if err != nil {
		t.Fatalf("ReadAsBytes error: %v", err)
	}
	if string(b) != string(content) {
		t.Errorf("File content got %q, want %q", string(b), string(content))
	}

	if IsDir(path) {
		t.Errorf("IsDir(%s) should be false for a file", path)
	}

	if !IsDir(dir) {
		t.Errorf("IsDir(%s) should be true for a directory", dir)
	}
}

func TestGetSize(t *testing.T) {
	dir := t.TempDir()
	content := "123456"
	filePath := createTempFile(t, dir, "file.txt", content)

	size, err := GetSize(filePath)
	if err != nil {
		t.Fatalf("GetSize error: %v", err)
	}
	if size != int64(len(content)) {
		t.Errorf("GetSize got %d, want %d", size, len(content))
	}
}

func TestDelAndDelAll(t *testing.T) {
	dir := t.TempDir()
	filePath := createTempFile(t, dir, "file.txt", "data")

	err := Del(filePath)
	if err != nil {
		t.Fatalf("Del error: %v", err)
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("file should be deleted")
	}

	// Test DelAll with directory
	subDir := filepath.Join(dir, "subdir")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	err = DelAll(subDir)
	if err != nil {
		t.Fatalf("DelAll error: %v", err)
	}
	if _, err := os.Stat(subDir); !os.IsNotExist(err) {
		t.Errorf("directory should be deleted")
	}
}

func TestCopy(t *testing.T) {
	dir := t.TempDir()
	src := createTempFile(t, dir, "src.txt", "copy content")
	dest := filepath.Join(dir, "dest.txt")

	err := Copy(src, dest)
	if err != nil {
		t.Fatalf("Copy error: %v", err)
	}

	b, err := ReadAsBytes(dest)
	if err != nil {
		t.Fatalf("ReadAsBytes error: %v", err)
	}
	if string(b) != "copy content" {
		t.Errorf("Copy content got %q, want %q", string(b), "copy content")
	}
}

func TestMoveOrRename(t *testing.T) {
	dir := t.TempDir()
	src := createTempFile(t, dir, "src.txt", "move me")
	dest := filepath.Join(dir, "dest.txt")

	err := MoveOrRename(src, dest)
	if err != nil {
		t.Fatalf("Move error: %v", err)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source file should be moved")
	}

	if _, err := os.Stat(dest); err != nil {
		t.Errorf("destination file should exist after move")
	}
}

func TestSplitFileName(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantName string
		wantExt  string
	}{
		{
			name:     "Simple file with extension",
			filename: "report.pdf",
			wantName: "report",
			wantExt:  "pdf",
		},
		{
			name:     "File with double extension",
			filename: "archive.tar.gz",
			wantName: "archive.tar",
			wantExt:  "gz",
		},
		{
			name:     "File without extension",
			filename: "README",
			wantName: "README",
			wantExt:  "",
		},
		{
			name:     "Hidden file without extension",
			filename: ".gitignore",
			wantName: ".gitignore",
			wantExt:  "",
		},
		{
			name:     "Hidden file with extension",
			filename: ".env.example",
			wantName: ".env",
			wantExt:  "example",
		},
		{
			name:     "File ending with dot",
			filename: "file.",
			wantName: "file.",
			wantExt:  "",
		},
		{
			name:     "File with multiple dots",
			filename: "a.b.c.d",
			wantName: "a.b.c",
			wantExt:  "d",
		},
		{
			name:     "Dot only",
			filename: ".",
			wantName: ".",
			wantExt:  "",
		},
		{
			name:     "Double dot",
			filename: "..",
			wantName: "..",
			wantExt:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotExt := SplitFileName(tt.filename)
			if gotName != tt.wantName || gotExt != tt.wantExt {
				t.Errorf("SplitFileName(%q) = (%q, %q), want (%q, %q)",
					tt.filename, gotName, gotExt, tt.wantName, tt.wantExt)
			}
		})
	}
}
