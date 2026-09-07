package filez

import (
	"fmt"
	"os"
	"syscall"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

// -----------------------------------------------------------------------------
// Creation & Deletion
// -----------------------------------------------------------------------------

// CreateParentDirs creates all parent directories for the given file path.
// It panics if any error occurs during directory creation.
func CreateParentDirs(file string) error {
	dir := Parent(file)
	return os.MkdirAll(dir, os.ModePerm)
}

// Remove deletes the file or directory at the given path. It returns `true` if
// the path was successfully removed, and `false` if the path did not exist or
// was not a directory. It panics for any other error.
func Remove(path string) error {
	err := os.Remove(path)
	if err != nil {
		e, ok := err.(*os.PathError)
		if ok && e.Err == syscall.ENOENT {
			return nil
		} else if ok && e.Err == syscall.ENOTDIR {
			return nil
		}
		return err
	}
	return nil
}

// RmTree recursively deletes the directory at the given path, along with all
// its contents. It is a convenience wrapper around `os.RemoveAll`.
func RmTree(dir string) error {
	return os.RemoveAll(dir)
}

// CreateTempDir creates a new temporary directory with a name based on the
// provided `name`. The directory name will be of the form `gox-tmp-<name>-*`.
// It returns the path to the created directory.
func CreateTempDir(name string) string {
	p := fmt.Sprintf("gox-tmp-%s-*", name)
	dir, err := os.MkdirTemp("", p)
	errorz.Check(err)
	return dir
}

// CreateTempFile creates a new temporary file with a name based on the provided
// `name`. The file name will be of the form `gox-tmp-<name>-*`. If `content` is
// not empty, it will be written to the file. It returns the path to the created file.
func CreateTempFile(name string, content []byte) string {
	p := fmt.Sprintf("gox-tmp-%s-*", name)
	file, err := os.CreateTemp("", p)
	errorz.Check(err)
	if len(content) > 0 {
		Write(file.Name(), content)
	}
	return file.Name()
}

// CreateFileFromPath creates a new file at the given path, including any necessary
// parent directories. It returns the created `*os.File`. It panics if any error
// occurs.
func CreateFileFromPath(path string) (*os.File, error) {
	if err := CreateParentDirs(path); err != nil {
		return nil, err
	}

	return os.Create(path)
}
