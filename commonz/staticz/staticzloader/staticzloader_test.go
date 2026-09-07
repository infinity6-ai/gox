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
		err := staticzloader.Walk(ctx, stzfiles.Name, func(path string, d fs.DirEntry, err error) error {
			count++
			f := expectedFiles[path]
			require.Equal(t, f.Dir, d.IsDir())
			if !d.IsDir() {
				info, errInfo := d.Info()
				require.NoError(t, errInfo)
				require.Equal(t, int64(len(f.Content)), info.Size())
			}
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, len(expectedFiles), count, "files count")
	})
}
