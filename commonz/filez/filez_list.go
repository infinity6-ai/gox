package filez

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

// -----------------------------------------------------------------------------
// Listing & Walking
// -----------------------------------------------------------------------------

// Walk traverses the directory tree starting from `base`, calling the provided
// `callback` function for each file and directory. The callback returns `true`
// to stop walking the current directory.
func Walk(base string, callback func(path string, f fs.DirEntry) error) error {
	return filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return callback(path, d)
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

type WalkLoaderEntry struct {
	Path  func() string
	Name  func() string
	IsDir func() bool
	Size  func() int64
	Open  func() (io.ReadCloser, error)
}

func WalkLoader(base string, callback func(entry WalkLoaderEntry) error) error {
	return Walk(base, func(path string, f fs.DirEntry) error {
		info, err := f.Info()
		if err != nil {
			return fmt.Errorf("error getting info for %s: %w", path, err)
		}
		return callback(WalkLoaderEntry{
			Path:  func() string { return path },
			Name:  f.Name,
			IsDir: f.IsDir,
			Size:  info.Size,
			Open: func() (io.ReadCloser, error) {
				if f.IsDir() {
					return nil, fmt.Errorf("cannot WalkLoad a directory: %s", path)
				}
				return os.Open(path)
			},
		})
	})
}
