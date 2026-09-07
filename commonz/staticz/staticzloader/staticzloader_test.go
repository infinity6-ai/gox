package staticzloader_test

import (
	"context"
	"io"
	"testing"

	"github.com/infinity6-ai/gox/commonz/filez"
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

	type F struct {
		Dir     bool
		Content string
	}
	expectedFiles := map[string]F{
		"stzfiles.txt":               {Content: "commonz\n"},
		"commonz":                    {Dir: true},
		"commonz/commonz-sample.txt": {Content: "commonz sample\n"},
	}

	t.Run("FindAll", func(t *testing.T) {
		var count int
		err := staticzloader.WalkLoader(ctx, stzfiles.Name, func(entry filez.WalkLoaderEntry) error {
			count++
			f := expectedFiles[entry.Path()]
			require.Equal(t, f.Dir, entry.IsDir())
			if !entry.IsDir() {
				r, err := entry.WalkLoad()
				require.NoError(t, err)
				data, err := io.ReadAll(r)
				require.NoError(t, err)
				require.Equal(t, f.Content, string(data))
				require.Equal(t, int64(len(f.Content)), entry.Size())
			}
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, len(expectedFiles), count, "files count")
	})
}
