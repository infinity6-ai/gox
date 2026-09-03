// Package filez provides a collection of utility functions for file and directory manipulation.
// These functions are designed to simplify common file system operations by providing robust,
// error-checked, and easy-to-use interfaces. The package includes functionalities for path
// manipulation, checking file existence and properties, creating and deleting files and directories,
// writing to and reading from files, and listing directory contents.
//
// The utilities in this package are built on top of the standard Go `os` and `path/filepath`
// packages, but with an emphasis on convenience and safety. For instance, many functions
// automatically create parent directories when writing files, and error handling is simplified
// through the use of `errorz.Check` to panic on unexpected errors, aligning with the project's
// error handling philosophy.
package filez

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/strconvz"

	"golang.org/x/sys/unix"
)

// -----------------------------------------------------------------------------
// Path Manipulation
// -----------------------------------------------------------------------------

// Parent returns the parent directory of the given path.
// It is a convenience wrapper around `filepath.Dir`.
func Parent(p string) string {
	return filepath.Dir(p)
}

// FindParent searches for a file with the given name by traversing up the directory
// tree starting from `startDir`. If the file is found, it returns the path to the
// directory containing the file and `true`. Otherwise, it returns an empty string
// and `false`.
func FindParent(filename string, startDir string) (string, bool) {
	dir := startDir
	if !path.IsAbs(dir) {
		d, err := filepath.Abs(dir)
		errorz.Check(err)
		dir = d
	}
	for {
		fullpath := filepath.Join(dir, filename)
		if FileExists(fullpath) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// -----------------------------------------------------------------------------
// Existence & Properties
// -----------------------------------------------------------------------------

// FileExists checks if a file or directory exists at the given path.
// It returns `true` if the path exists, and `false` if it does not.
// It panics for any error other than `fs.ErrNotExist`.
func FileExists(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false
		}
		errorz.Check(err)
	}
	return true
}

// IsDir checks if the given path is a directory. It returns `true` if the path
// is a directory, and `false` otherwise. If the path does not exist, it returns
// `false`. It panics for any other error.
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		errorz.Check(err)
	}
	return info.IsDir()
}

// IsDirEmpty checks if the directory at the given path is empty. It returns `true`
// if the directory is empty, and `false` otherwise. It panics if the path is not a
// directory or if any other error occurs.
func IsDirEmpty(path string) bool {
	dir, err := os.Open(path)
	errorz.Check(err)
	defer dir.Close()

	contents, err := dir.Readdirnames(1)
	if err != nil {
		if err == io.EOF {
			return true
		}
		errorz.Check(err)
	}

	return len(contents) == 0
}

// Stat returns the `fs.FileInfo` for the given path. If the path does not exist,
// it returns `nil`. It panics for any other error.
func Stat(path string) fs.FileInfo {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		errorz.Check(err)
	}
	return info
}

// Size returns the size of the file at the given path in bytes. If the file does
// not exist, it returns -1. It panics for any other error.
func Size(filePath string) int64 {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return -1
		}
		errorz.Check(err)
	}
	return info.Size()
}

// GetCreatedAt returns the creation time of the file at the given path. It uses
// `unix.Statx` to get the birth time of the file. If the birth time is not
// supported by the file system, it returns `nil`.
func GetCreatedAt(path string) *time.Time {
	var stat unix.Statx_t
	err := unix.Statx(unix.AT_FDCWD, path, unix.AT_STATX_SYNC_AS_STAT, unix.STATX_BTIME, &stat)
	errorz.Check(err)
	if stat.Mask&unix.STATX_BTIME == 0 {
		// not supported
		return nil
	}
	ret := time.Unix(stat.Btime.Sec, int64(stat.Btime.Nsec))
	return &ret
}

// GetUpdatedAt returns the last modification time of the file at the given path.
// It panics if any error occurs.
func GetUpdatedAt(path string) *time.Time {
	fi, err := os.Stat(path)
	errorz.Check(err)
	updatedTime := fi.ModTime()
	return &updatedTime
}

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

// -----------------------------------------------------------------------------
// Writing & Moving
// -----------------------------------------------------------------------------

// Write writes the given payload to a destination. The destination can be a file
// path or a file descriptor specified with the `@` prefix (e.g., `@1` for stdout).
// It panics if any error occurs.
func Write(dest string, payload []byte) {
	fd := -1
	if strings.HasPrefix(dest, "@") {
		fdStr := strings.TrimPrefix(dest, "@")
		fd = strconvz.MustParseNumber[int](fdStr)
	} else {
		file, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		errorz.Check(err)
		defer file.Close()
		fd = int(file.Fd())
	}
	_, err := syscall.Write(fd, payload)
	errorz.Check(err)
}

// WriteFile writes the given data to a file at the specified path. It creates
// parent directories if they do not exist. It panics if any error occurs.
func WriteFile(file string, data []byte) error {
	if err := CreateParentDirs(file); err != nil {
		return err
	}
	return os.WriteFile(file, data, os.ModePerm)
}

// WriteFromReader writes the content from the given `io.Reader` to a file at
// the specified path. It panics if any error occurs.
func WriteFromReader(path string, reader io.Reader) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	buf := make([]byte, 8*1024)
	_, err = io.CopyBuffer(file, reader, buf)
	return err
}

// Move moves a file from `fileFrom` to `fileTo`. It creates parent directories
// for the destination if they do not exist. It panics if any error occurs.
func Move(fileFrom string, fileTo string) error {
	if err := CreateParentDirs(fileTo); err != nil {
		return err
	}

	in, err := os.Open(fileFrom)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(fileTo)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return Remove(fileFrom)
}

// -----------------------------------------------------------------------------
// Reading
// -----------------------------------------------------------------------------

// ReadAllString reads all content from the given `io.Reader` and returns it as a
// string. It has a default limit of 10MB.
func ReadAllString(r io.Reader) string {
	return ReadAllStringLimited(r, 10*1024*1024)
}

// ReadAllStringLimited reads content from the given `io.Reader` up to a specified
// maximum number of bytes and returns it as a string.
func ReadAllStringLimited(r io.Reader, max int) string {
	var b bytes.Buffer
	b.ReadFrom(io.LimitReader(r, int64(max)))
	return b.String()
}

// ReadAllLimited reads all content from the given `io.Reader` up to a specified
// maximum number of bytes. It panics if the content size exceeds the limit.
func ReadAllLimited(r io.Reader, max int) blobz.Blob {
	body, err := io.ReadAll(io.LimitReader(r, int64(max+1)))
	errorz.Check(err)
	if len(body) > max {
		panic(fmt.Errorf("It is too large. Expected: %d, but was: %d", max, len(body)))
	}
	return blobz.New(body)
}

// TailFile reads the last `size` bytes of a file and returns them as a byte slice.
// If the file is smaller than `size`, it returns the entire file content.
func TailFile(file string, size int) []byte {
	f, err := os.Open(file)
	errorz.Check(err)
	defer f.Close()

	stat, err := f.Stat()
	errorz.Check(err)

	fileSize := stat.Size()
	if int64(size) > fileSize {
		size = int(fileSize)
	}

	if size == 0 {
		return []byte{}
	}

	buf := make([]byte, size)
	offset := fileSize - int64(size)
	_, err = f.ReadAt(buf, offset)
	errorz.Check(err)
	return buf
}

// ReadFile reads the content of a file up to a specified maximum number of bytes
// and returns it as a `blobz.Blob`. It panics if any error occurs during reading
// or if the file size exceeds the limit.
func ReadFile(file string, max int) blobz.Blob {
	f, err := os.Open(file)
	errorz.Check(err)
	defer f.Close()
	return ReadAllLimited(f, max)
}

// -----------------------------------------------------------------------------
// Listing & Walking
// -----------------------------------------------------------------------------

// Walk traverses the directory tree starting from `base`, calling the provided
// `callback` function for each file and directory. The callback returns `true`
// to stop walking the current directory.
func Walk(base string, callback func(path string, f fs.DirEntry) bool) {
	filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if !os.IsNotExist(err) {
				errorz.Check(err)
			}
			return nil
		}
		stop := callback(path, d)
		if stop {
			return filepath.SkipDir
		}
		return nil
	})
}

// Ls lists the contents of a directory and calls the provided `callback` function
// for each entry. The callback returns `true` to stop the listing.
func Ls(dir string, callback func(idx int, path string, f fs.DirEntry) (bool, error)) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	for idx, file := range files {
		path := filepath.Join(dir, file.Name())
		stop, err := callback(idx, path, file)
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
	}
	return nil
}

// DirList returns a list of file names in the given directory that match the
// provided regular expression. It panics if any error occurs.
func DirList(dir string, regex string) []string {
	ret := []string{}
	files, err := os.ReadDir(dir)
	errorz.Check(err)
	for _, file := range files {
		name := file.Name()
		match, err := regexp.MatchString(regex, name)
		errorz.Check(err)
		if match {
			ret = append(ret, file.Name())
		}
	}
	return ret
}

// DirListLimited is similar to `DirList`, but it returns at most `limit` number
// of matching file names.
func DirListLimited(dir string, regex string, limit int) []string {
	ret := []string{}
	files, err := os.ReadDir(dir)
	errorz.Check(err)
	for _, file := range files {
		name := file.Name()
		match, err := regexp.MatchString(regex, name)
		errorz.Check(err)
		if match {
			ret = append(ret, file.Name())
			if len(ret) >= limit {
				return ret
			}
		}
	}
	return ret
}
