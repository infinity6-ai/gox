package filez_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/stretchr/testify/assert"
)

func TestUnitRemove(t *testing.T) {
	// Test removing an existing file
	tmpFile, err := os.CreateTemp("", "test-remove")
	assert.NoError(t, err)
	filePath := tmpFile.Name()
	tmpFile.Close()

	assert.NoError(t, filez.Remove(filePath), "Should not return error for existing file")
	_, err = os.Stat(filePath)
	assert.True(t, os.IsNotExist(err), "File should not exist after removal")

	// Test removing a non-existent file
	assert.NoError(t, filez.Remove("/non/existent/file/path"), "Should not return error for non-existent file")

	// Test removing a non-existent file in an existing directory
	assert.NoError(t, filez.Remove(filepath.Join(os.TempDir(), "non-existent-file")), "Should not return error for non-existent file in existing dir")
}

func TestUnitMkdirAll(t *testing.T) {
	tmpDir := t.TempDir()
	newDirPath := filepath.Join(tmpDir, "new", "dir")

	err := filez.MkdirAll(newDirPath)
	assert.NoError(t, err)

	info, err := os.Stat(newDirPath)
	assert.NoError(t, err, "Directory should be created")
	assert.True(t, info.IsDir(), "Created path should be a directory")
}

func TestUnitCreateParentDirs(t *testing.T) {
	tmpDir := t.TempDir()
	newFilePath := filepath.Join(tmpDir, "new", "dir", "file.txt")

	err := filez.CreateParentDirs(newFilePath)
	assert.NoError(t, err)

	parentDir := filepath.Dir(newFilePath)
	info, err := os.Stat(parentDir)
	assert.NoError(t, err, "Parent directory should be created")
	assert.True(t, info.IsDir(), "Created path should be a directory")
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

func TestUnitCreateTempDir(t *testing.T) {
	dir := filez.CreateTempDir("my-test")
	defer os.RemoveAll(dir)

	info, err := os.Stat(dir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
	assert.Contains(t, filepath.Base(dir), "gox-tmp-my-test-")
}

func TestUnitCreateTempFile(t *testing.T) {
	content := []byte("temporary content")
	filePath := filez.CreateTempFile("my-test-file", content)
	defer os.Remove(filePath)

	info, err := os.Stat(filePath)
	assert.NoError(t, err)
	assert.False(t, info.IsDir())
	assert.Contains(t, filepath.Base(filePath), "gox-tmp-my-test-file-")

	readContent, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, content, readContent)
}
