package filez

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"golang.org/x/sys/unix"
)

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
