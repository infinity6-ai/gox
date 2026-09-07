package filez_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/stretchr/testify/assert"
)

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

func TestUnitSize(t *testing.T) {
	file := filepath.Join(t.TempDir(), "size-test.txt")
	content := "12345"
	err := os.WriteFile(file, []byte(content), 0644)
	assert.NoError(t, err)

	assert.Equal(t, int64(5), filez.Size(file))
	assert.Equal(t, int64(-1), filez.Size("non-existent-file"))
}
