package staticzloader_test

import (
	"context"
	"io"
	"io/fs"
	"testing"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/commonz/internal/stzfiles"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzentry"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzloader"
	"github.com/stretchr/testify/require"
)

func TestUnitGetCode(t *testing.T) {
	code := staticzloader.GetCode(stzfiles.Name)
	require.NotEmpty(t, code)
}

func TestUnitWalk(t *testing.T) {
	ctx := context.Background()

	expectedFiles := map[string]string{
		"stzfiles.txt":               "commonz\n",
		"commonz/commonz-sample.txt": "commonz sample\n",
	}

	t.Run("FindAll", func(t *testing.T) {
		var count int
		err := staticzloader.WalkV2(ctx, stzfiles.Name, func(entry staticzentry.Entry) error {
			count++
			f := expectedFiles[entry.Name()]
			r, err := entry.Open()
			require.NoError(t, err)
			data, err := io.ReadAll(r)
			require.NoError(t, err)
			require.Equal(t, f, string(data))
			require.Equal(t, int64(len(f)), entry.Size())
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, len(expectedFiles), count, "files count")
	})

	t.Run("SkipDir", func(t *testing.T) {
		var count int
		var f string
		err := staticzloader.Walk(ctx, stzfiles.Name, func(entry filez.WalkLoaderEntry) error {
			count++
			if entry.Path() == "commonz" {
				return fs.SkipDir
			}
			f = entry.Name()
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, "stzfiles.txt", f)
		require.Equal(t, 2, count, "files count")
	})
}
