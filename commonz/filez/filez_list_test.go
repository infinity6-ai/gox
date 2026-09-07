package filez_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"io"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitWalk(t *testing.T) {
	baseDir := t.TempDir()
	os.MkdirAll(filepath.Join(baseDir, "a", "b"), 0755)
	os.Create(filepath.Join(baseDir, "a", "file1.txt"))
	os.Create(filepath.Join(baseDir, "a", "b", "file2.txt"))

	var paths []string
	callback := func(path string, f fs.DirEntry) error {
		paths = append(paths, path)
		return nil
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

func TestUnitWalkLoader(t *testing.T) {
	baseDir := t.TempDir()

	// Create test directory structure
	file1Path := filepath.Join(baseDir, "file1.txt")
	subdir1Path := filepath.Join(baseDir, "subdir1")
	file2Path := filepath.Join(subdir1Path, "file2.txt")
	subdir2Path := filepath.Join(baseDir, "subdir2")
	file3Path := filepath.Join(subdir2Path, "file3.txt")

	err := os.WriteFile(file1Path, []byte("content of file1"), 0644)
	require.NoError(t, err)
	err = os.Mkdir(subdir1Path, 0755)
	require.NoError(t, err)
	err = os.WriteFile(file2Path, []byte("content of file2"), 0644)
	require.NoError(t, err)
	err = os.Mkdir(subdir2Path, 0755)
	require.NoError(t, err)
	err = os.WriteFile(file3Path, []byte("content of file3"), 0644)
	require.NoError(t, err)

	type expectedEntry struct {
		Name    string
		IsDir   bool
		Size    int64
		Content string // Only for files
	}

	expectedEntries := make(map[string]expectedEntry)
	// Base directory itself
	expectedEntries[filepath.ToSlash(baseDir)] = expectedEntry{
		Name:  filepath.Base(baseDir), // will be the temp dir name
		IsDir: true,
	}
	// Note: WalkLoader passes the full path, not relative.
	// We'll normalize to relative for map keys.
	expectedEntries[filepath.ToSlash("file1.txt")] = expectedEntry{
		Name:    "file1.txt",
		IsDir:   false,
		Size:    int64(len("content of file1")),
		Content: "content of file1",
	}
	expectedEntries[filepath.ToSlash("subdir1")] = expectedEntry{
		Name:  "subdir1",
		IsDir: true,
	}
	expectedEntries[filepath.ToSlash("subdir1/file2.txt")] = expectedEntry{
		Name:    "file2.txt",
		IsDir:   false,
		Size:    int64(len("content of file2")),
		Content: "content of file2",
	}
	expectedEntries[filepath.ToSlash("subdir2")] = expectedEntry{
		Name:  "subdir2",
		IsDir: true,
	}
	expectedEntries[filepath.ToSlash("subdir2/file3.txt")] = expectedEntry{
		Name:    "file3.txt",
		IsDir:   false,
		Size:    int64(len("content of file3")),
		Content: "content of file3",
	}

	actualEntries := make(map[string]struct {
		filez.WalkLoaderEntry
		Content string
	})

	t.Run("WalkLoader collects all entries correctly", func(t *testing.T) {
		err := filez.WalkLoader(baseDir, func(entry filez.WalkLoaderEntry) error {
			relPath, e := filepath.Rel(baseDir, entry.Path())
			require.NoError(t, e)

			// Special case for baseDir itself, relPath will be "."
			// Adjusting key to match the structure in expectedEntries
			key := filepath.ToSlash(relPath)
			if relPath == "." {
				key = filepath.ToSlash(baseDir)
			}
			
			var content string
			if !entry.IsDir() {
				rc, openErr := entry.WalkLoad()
				require.NoError(t, openErr)
				defer rc.Close()
				
				byteContent, readErr := io.ReadAll(rc)
				require.NoError(t, readErr)
				content = string(byteContent)
			}
			
			actualEntries[key] = struct {
				filez.WalkLoaderEntry
				Content string
			}{
				WalkLoaderEntry: entry,
				Content: content,
			}
			return nil
		})
		require.NoError(t, err)

		require.Len(t, actualEntries, len(expectedEntries), "should have correct number of entries")

		for expectedKey, expectedVal := range expectedEntries {
			t.Run("Check entry: "+expectedKey, func(t *testing.T) {
				actualVal, ok := actualEntries[expectedKey]
				require.True(t, ok, "expected entry %s not found in actual entries", expectedKey)

				// For the baseDir entry, the name will be the temp directory's name
				// We need to special case this for the assertion
				expectedName := expectedVal.Name
				if expectedKey == filepath.ToSlash(baseDir) {
					expectedName = filepath.Base(baseDir)
				}
				
				require.Equal(t, expectedName, actualVal.Name(), "Name mismatch for %s", expectedKey)
				require.Equal(t, expectedVal.IsDir, actualVal.IsDir(), "IsDir mismatch for %s", expectedKey)
				
				if !expectedVal.IsDir {
					require.Equal(t, expectedVal.Size, actualVal.Size(), "Size mismatch for file %s", expectedKey)
					require.Equal(t, expectedVal.Content, actualVal.Content, "Content mismatch for file %s", expectedKey)
				} else {
					// Verify that WalkLoad for directories returns an error (cannot open a directory)
					_, err := actualVal.WalkLoad()
					require.Error(t, err, "WalkLoad for directory %s should return an error", expectedKey)
				}
			})
		}
	})

	t.Run("WalkLoader handles error from callback", func(t *testing.T) {
		expectedErr := fmt.Errorf("simulated callback error")
		returnedErr := filez.WalkLoader(baseDir, func(entry filez.WalkLoaderEntry) error {
			// Return an error on the first entry to test error propagation
			return expectedErr
		})
		require.ErrorIs(t, returnedErr, expectedErr)
	})
}
