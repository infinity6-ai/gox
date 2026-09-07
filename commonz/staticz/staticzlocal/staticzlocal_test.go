package staticzlocal_test

import (
	"context"
	"io"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/commonz/internal/stzfiles"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzentry"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzlocal"
	"github.com/stretchr/testify/require"
)

func TestUnitBasic(t *testing.T) {
	dir, err := staticzlocal.LookupCurrentDir(stzfiles.Name)
	require.NoError(t, err)
	require.NotEmpty(t, dir)
	require.Equal(t, "commonz sample\n", filez.MustReadFile(filepath.Join(dir, "commonz", "commonz-sample.txt"), 256).String())
}

func TestUnitWalk(t *testing.T) {
	ctx := context.Background()

	expectedFiles := map[string]string{
		"stzfiles.txt":               "commonz\n",
		"commonz/commonz-sample.txt": "commonz sample\n",
		"commonz/commonz-s2.txt":     "commonz s2\n",
	}

	t.Run("FindAll", func(t *testing.T) {
		var count int
		err := staticzlocal.Walk(ctx, stzfiles.Name, func(entry staticzentry.Entry) error {
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

	t.Run("SkipAll", func(t *testing.T) {
		var count int
		var name string
		err := staticzlocal.Walk(ctx, stzfiles.Name, func(entry staticzentry.Entry) error {
			count++
			name = entry.Name()
			return fs.SkipAll
		})
		require.NoError(t, err)
		require.NotEmpty(t, name)
		require.NotEmpty(t, expectedFiles[name])
		require.Equal(t, 1, count, "files count")
	})
}
