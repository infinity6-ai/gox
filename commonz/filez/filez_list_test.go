package filez_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/stretchr/testify/assert"
)

func TestUnitWalk(t *testing.T) {
	baseDir := t.TempDir()
	os.MkdirAll(filepath.Join(baseDir, "a", "b"), 0755)
	os.Create(filepath.Join(baseDir, "a", "file1.txt"))
	os.Create(filepath.Join(baseDir, "a", "b", "file2.txt"))

	var paths []string
	callback := func(path string, f fs.DirEntry) bool {
		paths = append(paths, path)
		return false // continue walking
	}

	filez.Walk(baseDir, callback)

	// Normalize paths for comparison
	for i, p := range paths {
		paths[i], _ = filepath.Rel(baseDir, p)
		paths[i] = filepath.ToSlash(paths[i]) // for windows
	}

	expected := []string{".", "a", "a/b", "a/b/file2.txt", "a/file1.txt"}
	assert.ElementsMatch(t, expected, paths)
}

func TestUnitLs(t *testing.T) {
	dir := t.TempDir()
	os.Create(filepath.Join(dir, "file1.txt"))
	os.Create(filepath.Join(dir, "file2.log"))
	os.Mkdir(filepath.Join(dir, "subdir"), 0755)

	var entries []string
	callback := func(idx int, path string, f fs.DirEntry) (bool, error) {
		entries = append(entries, f.Name())
		return false, nil
	}

	err := filez.Ls(dir, callback)
	assert.NoError(t, err)

	assert.ElementsMatch(t, []string{"file1.txt", "file2.log", "subdir"}, entries)
}

func TestUnitDirList(t *testing.T) {
	dir := t.TempDir()
	os.Create(filepath.Join(dir, "test-1.log"))
	os.Create(filepath.Join(dir, "test-2.log"))
	os.Create(filepath.Join(dir, "other-1.txt"))

	// List all .log files
	logs := filez.DirList(dir, `^test-.*\.log$`)
	assert.ElementsMatch(t, []string{"test-1.log", "test-2.log"}, logs)

	// List all files
	all := filez.DirList(dir, `.*`)
	assert.ElementsMatch(t, []string{"test-1.log", "test-2.log", "other-1.txt"}, all)
}

func TestUnitDirListLimited(t *testing.T) {
	dir := t.TempDir()
	os.Create(filepath.Join(dir, "file-1.txt"))
	os.Create(filepath.Join(dir, "file-2.txt"))
	os.Create(filepath.Join(dir, "file-3.txt"))

	// Limit to 2 files
	limited := filez.DirListLimited(dir, `.*\.txt$`, 2)
	assert.Len(t, limited, 2)

	// Limit higher than number of files
	all := filez.DirListLimited(dir, `.*\.txt$`, 5)
	assert.Len(t, all, 3)
}
