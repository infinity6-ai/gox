package fsz_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil" // Added
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/fsz/fsz"
	"github.com/stretchr/testify/require"
)

func TestUnitFszFileProvider(t *testing.T) {
	tmpDir := filez.CreateTempDir("fsz-test")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// 1. Upload a file
	content := "hello world"
	filePath := filepath.Join(tmpDir, "test.txt")
	u, err := urlz.Parse("file://" + filePath)
	require.NoError(t, err)

	err = fsz.Upload(ctx, u, nil, strings.NewReader(content))
	require.NoError(t, err)

	// 2. Stat the file
	stat, err := fsz.Stat(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, stat)
	require.Equal(t, uint64(len(content)), stat.Size)
	require.Empty(t, stat.Md5)
	require.Empty(t, stat.Etag)

	// 3. Download the file
	err = fsz.Download(ctx, u, func(found bool, headers http.Header, reader io.Reader) error {
		require.True(t, found)
		require.NotNil(t, reader)

		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		require.Equal(t, content, string(data))
		return nil
	})
	require.NoError(t, err)

	// 4. Delete the file
	err = fsz.Delete(ctx, u)
	require.NoError(t, err)

	// 5. Stat again, should be nil
	stat, err = fsz.Stat(ctx, u)
	require.NoError(t, err)
	require.Nil(t, stat)
}

func TestUnitFszLs(t *testing.T) {
	tmpDir := filez.CreateTempDir("fsz-ls-test")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Create some files
	for i := 0; i < 5; i++ {
		filePath := filepath.Join(tmpDir, fmt.Sprintf("file%d.txt", i))
		err := os.WriteFile(filePath, fmt.Appendf(nil, "content %d", i), 0644)
		require.NoError(t, err)
	}

	u, err := urlz.Parse("file://" + tmpDir)
	require.NoError(t, err)

	paginator, err := fsz.Ls(ctx, u)
	require.NoError(t, err)
	defer paginator.Close()

	// Paginate with max=2
	stats1, err := paginator.Paginate(ctx, 2)
	require.NoError(t, err)
	require.Len(t, stats1, 2)

	// Paginate with max=2
	stats2, err := paginator.Paginate(ctx, 2)
	require.NoError(t, err)
	require.Len(t, stats2, 2)

	// Paginate with max=2 (should get the last one)
	stats3, err := paginator.Paginate(ctx, 2)
	require.NoError(t, err)
	require.Len(t, stats3, 1)

	// Paginate again, should be empty
	stats4, err := paginator.Paginate(ctx, 2)
	require.NoError(t, err)
	require.Len(t, stats4, 0)
}

func TestUnitDownloadNotFound(t *testing.T) {
	tmpDir := filez.CreateTempDir("fsz-notfound-test")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	filePath := filepath.Join(tmpDir, "nonexistent.txt")
	u, err := urlz.Parse("file://" + filePath)
	require.NoError(t, err)

	err = fsz.Download(ctx, u, func(found bool, headers http.Header, reader io.Reader) error {
		require.False(t, found)
		require.Nil(t, headers)
		require.Nil(t, reader)
		return nil
	})
	require.NoError(t, err)
}

func TestUnitStatNotFound(t *testing.T) {
	tmpDir := filez.CreateTempDir("fsz-stat-notfound-test")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	filePath := filepath.Join(tmpDir, "nonexistent.txt")
	u, err := urlz.Parse("file://" + filePath)
	require.NoError(t, err)

	stat, err := fsz.Stat(ctx, u)
	require.NoError(t, err)
	require.Nil(t, stat)
}

func TestUnitDownloadCallbackReader(t *testing.T) {
	tmpDir := filez.CreateTempDir("fsz-download-callback-test")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	content := "some test data"
	filePath := filepath.Join(tmpDir, "data.txt")
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	u, err := urlz.Parse("file://" + filePath)
	require.NoError(t, err)

	var downloadedData bytes.Buffer
	err = fsz.Download(ctx, u, func(found bool, headers http.Header, reader io.Reader) error {
		require.True(t, found)
		_, err := io.Copy(&downloadedData, reader)
		require.NoError(t, err)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, content, downloadedData.String())
}

func TestUnitFszCopy(t *testing.T) {
	tmpDir := filez.CreateTempDir("fsz-copy-test")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	type testScenario struct {
		name          string
		srcContent    string
		srcFilename   string
		destFilename  string
		expectError   bool
		expectedError string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		srcPath := filepath.Join(tmpDir, s.srcFilename)
		destPath := filepath.Join(tmpDir, s.destFilename)

		// Create source file if content is provided
		if s.srcContent != "" {
			err := os.WriteFile(srcPath, []byte(s.srcContent), 0644)
			require.NoError(t, err)
		}

		srcURL, err := urlz.Parse("file://" + srcPath)
		require.NoError(t, err)
		destURL, err := urlz.Parse("file://" + destPath)
		require.NoError(t, err)

		err = fsz.Copy(ctx, srcURL, destURL)

		if s.expectError {
			require.Error(t, err)
			if s.expectedError != "" {
				require.Contains(t, err.Error(), s.expectedError)
			}
			// Ensure destination file does not exist if copy failed as expected
			_, statErr := os.Stat(destPath)
			require.True(t, os.IsNotExist(statErr), "Destination file should not exist if copy failed")
		} else {
			require.NoError(t, err)

			// Verify destination file exists and content is correct
			destContent, err := os.ReadFile(destPath)
			require.NoError(t, err)
			require.Equal(t, s.srcContent, string(destContent))

			// Verify Stat on destination
			stat, err := fsz.Stat(ctx, destURL)
			require.NoError(t, err)
			require.NotNil(t, stat)
			require.Equal(t, uint64(len(s.srcContent)), stat.Size)
		}
	}

	t.Run("Copy existing file", func(t *testing.T) {
		check(t, testScenario{
			name:         "Copy existing file",
			srcContent:   "content for source file",
			srcFilename:  "src.txt",
			destFilename: "dest.txt",
			expectError:  false,
		})
	})

	t.Run("Copy non-existent source file", func(t *testing.T) {
		check(t, testScenario{
			name:          "Copy non-existent source file",
			srcContent:    "", // Means source file is not created
			srcFilename:   "nonexistent_src.txt",
			destFilename:  "nonexistent_dest.txt",
			expectError:   true,
			expectedError: "failed to open source file",
		})
	})

	t.Run("Copy to a new directory", func(t *testing.T) {
		check(t, testScenario{
			name:         "Copy to a new directory",
			srcContent:   "content for another source",
			srcFilename:  "src2.txt",
			destFilename: "newdir/dest2.txt",
			expectError:  false,
		})
	})

	t.Run("Overwrite existing destination file", func(t *testing.T) {
		// Create a destination file with different content first
		existingDestPath := filepath.Join(tmpDir, "overwrite_dest.txt")
		err := os.WriteFile(existingDestPath, []byte("original content"), 0644)
		require.NoError(t, err)

		check(t, testScenario{
			name:         "Overwrite existing destination file",
			srcContent:   "new content to overwrite",
			srcFilename:  "overwrite_src.txt",
			destFilename: "overwrite_dest.txt",
			expectError:  false,
		})
		// Verify the content is overwritten
		overwrittenContent, err := os.ReadFile(existingDestPath)
		require.NoError(t, err)
		require.Equal(t, "new content to overwrite", string(overwrittenContent))
	})
}

func TestUnitFszFind(t *testing.T) {
	tmpDir := filez.CreateTempDir("fsz-find-test")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Create some files
	files := []string{
		"a/b/c.txt",
		"a/b/d.txt",
		"a/e.txt",
		"f.txt",
		"g/h/i.log",
		"g/j.txt",
	}
	for _, f := range files {
		filePath := filepath.Join(tmpDir, f)
		err := filez.CreateParentDirs(filePath)
		require.NoError(t, err)
		err = os.WriteFile(filePath, []byte("content"), 0644)
		require.NoError(t, err)
	}

	u, err := urlz.Parse("file://" + tmpDir)
	require.NoError(t, err)

	paginator, err := fsz.Find(ctx, u)
	require.NoError(t, err)
	defer paginator.Close()

	var foundFiles int
	for {
		stats, err := paginator.Paginate(ctx, 2)
		require.NoError(t, err)
		if len(stats) == 0 {
			break
		}
		foundFiles += len(stats)
	}

	require.Equal(t, len(files), foundFiles)
}

func TestUnitFileFsLsNonExistentDir(t *testing.T) {
	// Create a path that is guaranteed not to exist
	nonExistentDirPath := filepath.Join(t.TempDir(), "non-existent-dir")
	nonExistentDirUrl := urlz.MustParse("file://" + filepath.ToSlash(nonExistentDirPath))

	paginator, err := fsz.Ls(context.Background(), nonExistentDirUrl)

	require.NoError(t, err)
	require.NotNil(t, paginator)

	stats, paginateErr := paginator.Paginate(context.Background(), 10)
	require.NoError(t, paginateErr)
	require.Empty(t, stats)
	require.NoError(t, paginator.Close())
}

func TestUnitFileFsFindNonExistentDir(t *testing.T) {
	// Create a path that is guaranteed not to exist
	nonExistentDirPath := filepath.Join(t.TempDir(), "non-existent-dir")
	nonExistentDirUrl := urlz.MustParse("file://" + filepath.ToSlash(nonExistentDirPath))

	paginator, err := fsz.Find(context.Background(), nonExistentDirUrl)

	require.NoError(t, err)   // Should not return an error
	require.NotNil(t, paginator) // Should return a non-nil paginator (fileFindPaginator)

	// Ensure it returns no results and no error on Paginate
	stats, paginateErr := paginator.Paginate(context.Background(), 10)
	require.NoError(t, paginateErr)
	require.Empty(t, stats)
	require.NoError(t, paginator.Close())
}

func TestUnitFileFsLsEmptyDir(t *testing.T) {
	// Create an empty directory
	emptyDirPath := t.TempDir()
	emptyDirUrl := urlz.MustParse("file://" + filepath.ToSlash(emptyDirPath))

	paginator, err := fsz.Ls(context.Background(), emptyDirUrl)

	require.NoError(t, err)
	require.NotNil(t, paginator)

	stats, paginateErr := paginator.Paginate(context.Background(), 10)
	require.NoError(t, paginateErr)
	require.Empty(t, stats)
	require.NoError(t, paginator.Close())
}

func TestUnitFileFsFindEmptyDir(t *testing.T) {
	// Create an empty directory
	emptyDirPath := t.TempDir()
	emptyDirUrl := urlz.MustParse("file://" + filepath.ToSlash(emptyDirPath))

	paginator, err := fsz.Find(context.Background(), emptyDirUrl)

	require.NoError(t, err)
	require.NotNil(t, paginator)

	stats, paginateErr := paginator.Paginate(context.Background(), 10)
	require.NoError(t, paginateErr)
	require.Empty(t, stats)
	require.NoError(t, paginator.Close())
}

func TestUnitFileFsRmTree(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory with content
	dirToRemove := filepath.Join(tmpDir, "dir-to-remove")
	require.NoError(t, os.Mkdir(dirToRemove, 0755))
	require.NoError(t, ioutil.WriteFile(filepath.Join(dirToRemove, "file1.txt"), []byte("content"), 0644))
	subDir := filepath.Join(dirToRemove, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0755))
	require.NoError(t, ioutil.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("content"), 0644))

	dirUrl := urlz.MustParse("file://" + filepath.ToSlash(dirToRemove))

	err := fsz.RmTree(context.Background(), dirUrl)
	require.NoError(t, err)

	_, err = os.Stat(dirToRemove)
	require.True(t, os.IsNotExist(err), "directory should not exist after RmTree")
}


func TestUnitFileFsMoveFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a source file
	srcPath := filepath.Join(tmpDir, "src.txt")
	require.NoError(t, ioutil.WriteFile(srcPath, []byte("content"), 0644))

	destPath := filepath.Join(tmpDir, "dest.txt")
	srcUrl := urlz.MustParse("file://" + filepath.ToSlash(srcPath))
	destUrl := urlz.MustParse("file://" + filepath.ToSlash(destPath))

	err := fsz.Move(context.Background(), srcUrl, destUrl)
	require.NoError(t, err)

	_, err = os.Stat(srcPath)
	require.True(t, os.IsNotExist(err), "source file should not exist after move")

	_, err = os.Stat(destPath)
	require.NoError(t, err, "destination file should exist after move")

	content, err := ioutil.ReadFile(destPath)
	require.NoError(t, err)
	require.Equal(t, "content", string(content))
}


func TestUnitFileFsMoveDirFails(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a source directory with content
	srcDirPath := filepath.Join(tmpDir, "src-dir")
	require.NoError(t, os.Mkdir(srcDirPath, 0755))
	require.NoError(t, ioutil.WriteFile(filepath.Join(srcDirPath, "file.txt"), []byte("content"), 0644))

	destDirPath := filepath.Join(tmpDir, "dest-dir")
	srcUrl := urlz.MustParse("file://" + filepath.ToSlash(srcDirPath))
	destUrl := urlz.MustParse("file://" + filepath.ToSlash(destDirPath))

	err := fsz.Move(context.Background(), srcUrl, destUrl)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Move is only for files")

	// Assert source directory still exists
	_, err = os.Stat(srcDirPath)
	require.NoError(t, err, "source directory should still exist after failed move")
}
