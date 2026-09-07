package staticzloader_test

import (
	"context"
	"io/fs"
	"testing"

	"github.com/infinity6-ai/gox/commonz/internal/stzfiles"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzloader"
	"github.com/stretchr/testify/require"
)

func TestUnitGetCode(t *testing.T) {
	code := staticzloader.GetCode(stzfiles.Name)
	require.NotEmpty(t, code)
}

func TestUnitWalk(t *testing.T) {
	ctx := context.Background()

	t.Run("FindsExpectedFile", func(t *testing.T) {
		var found bool
		walkFn := func(path string, d fs.DirEntry, err error) error {
			require.NoError(t, err)
			if path == "stzfiles.txt" {
				found = true
				require.False(t, d.IsDir())
				info, errInfo := d.Info()
				require.NoError(t, errInfo)
				require.Equal(t, int64(8), info.Size()) // "commonz\n" is 8 bytes
			}
			return nil
		}

		err := staticzloader.Walk(ctx, stzfiles.Name, walkFn)
		require.NoError(t, err)
		require.True(t, found, "expected file 'stzfiles.txt' was not found")
	})

	t.Run("CollectsAllFiles", func(t *testing.T) {
		paths := make(map[string]fs.DirEntry)
		walkFn := func(path string, d fs.DirEntry, err error) error {
			require.NoError(t, err)
			paths[path] = d
			return nil
		}

		err := staticzloader.Walk(ctx, stzfiles.Name, walkFn)
		require.NoError(t, err)

		require.Contains(t, paths, "stzfiles.txt", "archive should contain stzfiles.txt")

		// Check that it's the only file, if no directories are present.
		// This makes the test more robust.
		if _, hasDir := paths["./"]; !hasDir {
			require.Len(t, paths, 1, "should only contain one file")
		}
	})
}
