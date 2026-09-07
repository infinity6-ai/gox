package filez

import (
	"path"
	"path/filepath"

	"github.com/infinity6-ai/gox/commonz/errorz"
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
