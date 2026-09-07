package staticz_test

import (
	"context"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/infinity6-ai/gox/commonz/internal/stzfiles"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/staticz"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzentry"
	"github.com/stretchr/testify/require"
)

func TestUnitWalkAndLookup(t *testing.T) {
	ctx := context.Background()

	expectedFiles := map[string]string{
		"stzfiles.txt":               "commonz\n",
		"commonz/commonz-sample.txt": "commonz sample\n",
		"commonz/commonz-s2.txt":     "commonz s2\n",
	}

	t.Run("FindAll", func(t *testing.T) {
		var count int
		err := staticz.Walk(ctx, stzfiles.Name, func(entry staticzentry.Entry) error {
			count++
			f := expectedFiles[entry.Name().String()]
			r, err := entry.Open()
			require.NoError(t, err)
			defer r.Close()
			data, err := io.ReadAll(r)
			require.NoError(t, err)
			require.Equal(t, f, string(data))
			require.Equal(t, int64(len(f)), entry.Size())
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, len(expectedFiles), count, "files count")
	})

	t.Run("SkipAll", func(t *testing.T) {
		var count int
		var name string
		err := staticz.Walk(ctx, stzfiles.Name, func(entry staticzentry.Entry) error {
			count++
			name = entry.Name().String()
			return fs.SkipAll
		})
		require.NoError(t, err)
		require.NotEmpty(t, name)
		require.NotEmpty(t, expectedFiles[name])
		require.Equal(t, 1, count, "files count")
	})

	t.Run("Lookup found", func(t *testing.T) {
		entry, err := staticz.Lookup(ctx, stzfiles.Name, pathz.MustParse("commonz/commonz-sample.txt"))
		require.NoError(t, err)
		require.NotNil(t, entry)
		require.Equal(t, "commonz sample\n", expectedFiles[entry.Name().String()])
	})

	t.Run("Lookup not found", func(t *testing.T) {
		entry, err := staticz.Lookup(ctx, stzfiles.Name, pathz.MustParse("commonz/notfound.txt"))
		require.NoError(t, err)
		require.Nil(t, entry)
	})
}

func TestUnitExtractTo(t *testing.T) {
	ctx := context.Background()

	expectedFiles := map[string]string{
		"stzfiles.txt":               "commonz\n",
		"commonz/commonz-sample.txt": "commonz sample\n",
		"commonz/commonz-s2.txt":     "commonz s2\n",
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "staticz-extract-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir) // Clean up after the test

	destPath := pathz.MustParse(tempDir)

	// Extract files
	err = staticz.ExtractTo(ctx, destPath, stzfiles.Name)
	require.NoError(t, err)

	// Verify extracted files
	for fileName, expectedContent := range expectedFiles {
		filePath := destPath.MustJoin(pathz.MustParse(fileName))
		fileContent, err := os.ReadFile(filePath.String())
		require.NoError(t, err, "Failed to read extracted file: %s", filePath.String())
		require.Equal(t, expectedContent, string(fileContent), "Content mismatch for file: %s", filePath.String())

		fileInfo, err := os.Stat(filePath.String())
		require.NoError(t, err)
		require.Equal(t, int64(len(expectedContent)), fileInfo.Size(), "Size mismatch for file: %s", filePath.String())
	}

	// Verify that an attempt to extract to an invalid path fails
	invalidDestPath := pathz.MustParse("/nonexistent/path/that/should/fail")
	err = staticz.ExtractTo(ctx, invalidDestPath, stzfiles.Name)
	require.Error(t, err)
}

