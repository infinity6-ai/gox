package filez_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/stretchr/testify/assert"
)

func TestUnitParent(t *testing.T) {
	assert.Equal(t, "/path/to", filez.Parent("/path/to/file"))
	assert.Equal(t, "/", filez.Parent("/"))
	assert.Equal(t, ".", filez.Parent(""))
	assert.Equal(t, ".", filez.Parent("any.txt"))
	assert.Equal(t, ".", filez.Parent("any"))
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
