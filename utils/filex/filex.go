package filex

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// permission
const (
	OwnerWrite = 0644
	Full       = os.ModePerm // 0777
)

// Join joins any number of path elements into a single path,
// using the operating system's path separator.
//
// It cleans the result by removing redundant separators
// and resolving "." and ".." elements.
//
// Example:
//
//	p := Join("/tmp", "data", "file.txt")
//	// On Unix:    "/tmp/data/file.txt"
//	// On Windows: "C:\\tmp\\data\\file.txt"
func Join(paths ...string) string {
	return filepath.Join(paths...)
}

// ReadAsBytes reads the entire content of the file specified by `path`
// and returns it as a byte slice.
//
// Returns an error if the file cannot be read.
func ReadAsBytes(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ReadAsString reads the entire content of the file specified by `path`
// and returns it as a string.
//
// Returns an error if the file cannot be read.
func ReadAsString(path string) (string, error) {
	data, err := ReadAsBytes(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadAsLines reads the content of the file specified by `path`
// and returns it as a slice of strings, where each element is a line.
//
// Returns an error if the file cannot be opened or read.
func ReadAsLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// CreateDir creates a directory named by `path`, along with any necessary parent directories.
// It returns an error if the directory cannot be created.
//
// Example:
//
//	err := CreateDir("/tmp/myapp/data")
//	if err != nil {
//	    log.Fatal(err)
//	}
func CreateDir(path string) error {
	return os.MkdirAll(path, Full)
}

// CreateDirTemp creates a new temporary directory in the directory `path`
// (or the system default temp directory if path is empty) with a name beginning with `prefix`.
// It returns the path of the new directory or an error.
//
// Note:
//   - The caller is responsible for removing the temporary directory after use,
//     typically with `defer os.RemoveAll(dir)` to avoid leaving temp files behind.
//
// Example:
//
//	dir, err := CreateDirTemp("myapp-")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer os.RemoveAll(dir) // clean up after use
func CreateDirTemp(prefix string) (string, error) {
	return os.MkdirTemp("", prefix)
}

// SaveFile saves the given content to a file specified by dirPath and filename.
// It creates the directory if it does not exist.
//
// Parameters:
//   - dirPath: the directory path where the file should be saved.
//   - filename: the name of the file to save.
//   - content: the byte slice content to write into the file.
//
// Returns an error if the directory cannot be created or the file cannot be written.
func SaveFile(dirPath, filename string, content []byte) error {
	// Create the directory if it doesn't exist
	err := CreateDir(dirPath)
	if err != nil {
		return err
	}

	// Combine directory and filename to get full file path
	fullPath := filepath.Join(dirPath, filename)

	// Write content to file
	return os.WriteFile(fullPath, content, OwnerWrite)
}

// IsDir checks if a given path is a directory.
func IsDir(path string) bool {
	inf, err := os.Stat(path)
	if err != nil {
		return false
	}

	return inf.IsDir()
}

func GetSize(path string) (int64, error) {
	inf, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return inf.Size(), nil
}

// Copy copies the contents of the source file to the destination file.
//
// It reads the entire source file into memory and writes it to the destination.
// The destination file will be created with permission 0644.
// If the destination file exists, it will be overwritten.
//
// Note:
//   - This approach may not be efficient for very large files.
//
// Example:
//
//	err := Copy("input.txt", "backup/input.txt")
//	if err != nil {
//	    log.Fatal(err)
//	}
func Copy(src, dest string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, input, OwnerWrite)
}

// MoveOrRename renames (moves) a file or directory from src to dest.
//
// If src and dest are in the same filesystem, it will rename or move instantly.
// If src and dest are in different filesystems, this will fail with an error.
//
// Note:
//   - This function uses os.Rename under the hood.
//   - To move files across different filesystems, you need to copy the file and delete the original manually.
//
// Example:
//
//	err := Move("data.txt", "/tmp/data.txt")
//	if err != nil {
//	    log.Fatal(err)
//	}
func MoveOrRename(src, dest string) error {
	return os.Rename(src, dest)
}

// Del deletes the file or (empty) directory at the specified path.
//
// It uses os.Remove, so:
//   - If the target is a file, it will be deleted.
//   - If the target is an empty directory, it will be removed.
//   - If the directory is not empty, this will return an error.
//
// Example:
//
//	err := Del("data.txt")
//	if err != nil {
//	    log.Fatal(err)
//	}
func Del(path string) error {
	return os.Remove(path)
}

// DelAll deletes the specified file or directory and all its contents.
//
// If the path is a file, it will be removed.
// If the path is a directory, it and all nested files and subdirectories
// will be deleted recursively.
//
// This is similar to running `rm -rf` in Unix.
//
// Example:
//
//	err := DelAll("/tmp/data")
//	if err != nil {
//	    log.Fatal(err)
//	}
func DelAll(path string) error {
	return os.RemoveAll(path)
}

// SplitFileName splits a filename into name and extension parts.
// If there is no extension, ext will be "".
//
// Example:
//
//	"report.pdf"       -> "report", "pdf"
//	"archive.tar.gz"   -> "archive.tar", "gz"
//	"README"           -> "README", ""
//	".gitignore"       -> ".gitignore", ""
func SplitFileName(filename string) (name string, ext string) {
	lastDot := strings.LastIndex(filename, ".")
	if lastDot == -1 || lastDot == 0 || lastDot == len(filename)-1 {
		// No extension found (or hidden file like .gitignore)
		return filename, ""
	}
	return filename[:lastDot], filename[lastDot+1:]
}
