package filez_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.code.infinity6.ai/platform/util/filez"
)

func TestUnitParent(t *testing.T) {
	assert.Equal(t, "/path/to", filez.Parent("/path/to/file"))
	assert.Equal(t, "/", filez.Parent("/"))
	assert.Equal(t, ".", filez.Parent(""))
	assert.Equal(t, ".", filez.Parent("any.txt"))
	assert.Equal(t, ".", filez.Parent("any"))
}

func TestUnitFromUrl(t *testing.T) {
	assert.Equal(t, "/path/to/file", filez.FromUrl("file:///path/to/file"))
	assert.Equal(t, "/path/to/file", filez.FromUrl("file:/path/to/file"))
	assert.Equal(t, "path/to/file", filez.FromUrl("path/to/file"))
}

func TestUnitRemove(t *testing.T) {
	// Test removing an existing file
	tmpFile, err := os.CreateTemp("", "test-remove")
	assert.NoError(t, err)
	filePath := tmpFile.Name()
	tmpFile.Close()

	assert.True(t, filez.Remove(filePath), "Should return true for existing file")
	_, err = os.Stat(filePath)
	assert.True(t, os.IsNotExist(err), "File should not exist after removal")

	// Test removing a non-existent file
	assert.False(t, filez.Remove("/non/existent/file/path"), "Should return false for non-existent file")

	// Test removing a non-existent file in an existing directory
	assert.False(t, filez.Remove(filepath.Join(os.TempDir(), "non-existent-file")), "Should return false for non-existent file in existing dir")
}

func TestUnitCreateParentDirs(t *testing.T) {
	tmpDir := t.TempDir()
	newFilePath := filepath.Join(tmpDir, "new", "dir", "file.txt")

	filez.CreateParentDirs(newFilePath)

	parentDir := filepath.Dir(newFilePath)
	info, err := os.Stat(parentDir)
	assert.NoError(t, err, "Parent directory should be created")
	assert.True(t, info.IsDir(), "Created path should be a directory")
}

func TestUnitFileExists(t *testing.T) {
	// Test with an existing file
	tmpFile, err := os.CreateTemp("", "test-exists")
	assert.NoError(t, err)
	filePath := tmpFile.Name()
	defer os.Remove(filePath)
	tmpFile.Close()

	assert.True(t, filez.FileExists(filePath), "Should return true for an existing file")

	// Test with a non-existent file
	assert.False(t, filez.FileExists("/non/existent/file/path"), "Should return false for a non-existent file")

	// Test with a directory
	tmpDir := t.TempDir()
	assert.True(t, filez.FileExists(tmpDir), "Should return true for a directory")
}

func TestUnitWrite(t *testing.T) {
	// Test writing to a file path
	tmpFile := filepath.Join(t.TempDir(), "test-write.txt")
	payload1 := []byte("hello")
	filez.Write(tmpFile, payload1)

	content, err := os.ReadFile(tmpFile)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(content))

	// Test appending to the same file
	payload2 := []byte(" world")
	filez.Write(tmpFile, payload2)

	content, err = os.ReadFile(tmpFile)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", string(content))
}

func TestUnitCreateTempSocket(t *testing.T) {
	socketPath := filez.CreateTempSocket()
	defer os.RemoveAll(filepath.Dir(socketPath))

	assert.Contains(t, socketPath, "i6go-platform-tmp-xz-")
	assert.True(t, strings.HasSuffix(socketPath, "xz.sock"))

	// Check if parent dir was created
	dirInfo, err := os.Stat(filepath.Dir(socketPath))
	assert.NoError(t, err)
	assert.True(t, dirInfo.IsDir())
}

func TestUnitRmTree(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	err := os.Mkdir(subDir, 0755)
	assert.NoError(t, err)

	tmpFile, err := os.Create(filepath.Join(subDir, "file.txt"))
	assert.NoError(t, err)
	tmpFile.Close()

	err = filez.RmTree(tmpDir)
	assert.NoError(t, err)

	_, err = os.Stat(tmpDir)
	assert.True(t, os.IsNotExist(err), "Directory should be removed")
}

func TestUnitFindParent(t *testing.T) {
	// Setup directory structure
	baseDir := t.TempDir()
	dir1 := filepath.Join(baseDir, "dir1")
	dir2 := filepath.Join(dir1, "dir2")
	dir3 := filepath.Join(dir2, "dir3")
	os.MkdirAll(dir3, 0755)

	// Create the target file in dir2
	targetFile := "my-file.txt"
	f, err := os.Create(filepath.Join(dir2, targetFile))
	assert.NoError(t, err)
	f.Close()

	// Search from a deeper directory
	foundDir, found := filez.FindParent(targetFile, dir3)
	assert.True(t, found, "Should find the parent directory")
	assert.Equal(t, dir2, foundDir, "Should find the correct parent directory")

	// Search from the directory that contains it
	foundDir, found = filez.FindParent(targetFile, dir2)
	assert.True(t, found, "Should find the parent directory")
	assert.Equal(t, dir2, foundDir, "Should find the correct parent directory")

	// Search for a non-existent file
	_, found = filez.FindParent("non-existent-file.txt", dir3)
	assert.False(t, found, "Should not find a non-existent file")
}

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
	callback := func(idx int, path string, f fs.DirEntry) bool {
		entries = append(entries, f.Name())
		return false
	}

	filez.Ls(dir, callback)

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

func TestUnitCreateTempDir(t *testing.T) {
	dir := filez.CreateTempDir("my-test")
	defer os.RemoveAll(dir)

	info, err := os.Stat(dir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
	assert.Contains(t, filepath.Base(dir), "i6go-platform-tmp-my-test-")
}

func TestUnitCreateTempFile(t *testing.T) {
	content := []byte("temporary content")
	filePath := filez.CreateTempFile("my-test-file", content)
	defer os.Remove(filePath)

	info, err := os.Stat(filePath)
	assert.NoError(t, err)
	assert.False(t, info.IsDir())
	assert.Contains(t, filepath.Base(filePath), "i6go-platform-tmp-my-test-file-")

	readContent, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, content, readContent)
}

func TestUnitIsDir(t *testing.T) {
	dir := t.TempDir()
	file, _ := os.CreateTemp(dir, "file")

	assert.True(t, filez.IsDir(dir))
	assert.False(t, filez.IsDir(file.Name()))
	assert.False(t, filez.IsDir("non-existent-path"))
}

func TestUnitIsDirEmpty(t *testing.T) {
	// Test with an empty directory
	emptyDir := t.TempDir()
	assert.True(t, filez.IsDirEmpty(emptyDir), "Should be empty")

	// Test with a non-empty directory
	nonEmptyDir := t.TempDir()
	f, err := os.Create(filepath.Join(nonEmptyDir, "file.txt"))
	assert.NoError(t, err)
	f.Close()
	assert.False(t, filez.IsDirEmpty(nonEmptyDir), "Should not be empty")
}

func TestUnitMove(t *testing.T) {
	dir := t.TempDir()
	fileFrom := filepath.Join(dir, "from.txt")
	fileTo := filepath.Join(dir, "newdir", "to.txt")
	content := []byte("move me")

	err := os.WriteFile(fileFrom, content, 0644)
	assert.NoError(t, err)

	filez.Move(fileFrom, fileTo)

	// Check if source is gone
	_, err = os.Stat(fileFrom)
	assert.True(t, os.IsNotExist(err), "Source file should be removed")

	// Check if destination exists with correct content
	readContent, err := os.ReadFile(fileTo)
	assert.NoError(t, err)
	assert.Equal(t, content, readContent)
}

func TestUnitSize(t *testing.T) {
	file := filepath.Join(t.TempDir(), "size-test.txt")
	content := "12345"
	err := os.WriteFile(file, []byte(content), 0644)
	assert.NoError(t, err)

	assert.Equal(t, int64(5), filez.Size(file))
	assert.Equal(t, int64(-1), filez.Size("non-existent-file"))
}
