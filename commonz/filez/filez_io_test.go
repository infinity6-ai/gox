package filez_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/stretchr/testify/assert"
)

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

func TestUnitMove(t *testing.T) {
	dir := t.TempDir()
	fileFrom := filepath.Join(dir, "from.txt")
	fileTo := filepath.Join(dir, "newdir", "to.txt")
	content := []byte("move me")

	err := os.WriteFile(fileFrom, content, 0644)
	assert.NoError(t, err)

	err = filez.Move(fileFrom, fileTo)
	assert.NoError(t, err)

	// Check if source is gone
	_, err = os.Stat(fileFrom)
	assert.True(t, os.IsNotExist(err), "Source file should be removed")

	// Check if destination exists with correct content
	readContent, err := os.ReadFile(fileTo)
	assert.NoError(t, err)
	assert.Equal(t, content, readContent)
}
