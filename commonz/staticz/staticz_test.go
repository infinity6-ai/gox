package staticz_test

import (
	"context"
	"io"
	"io/fs"
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
