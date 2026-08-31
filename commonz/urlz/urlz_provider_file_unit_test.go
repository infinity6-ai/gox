package urlz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/stretchr/testify/require"
)

func TestUnitParseFile(t *testing.T) {
	t.Run("file", func(t *testing.T) {
		p, err := pathz.Parse("/tmp/x")
		require.NoError(t, err)

		// Remaking the test a bit since String() is not stable for file
		rawUrl := "file:///tmp/x"
		got, err := urlz.Parse(rawUrl)
		require.NoError(t, err)
		require.Equal(t, &urlz.Url{Scheme: "file", Path: p}, got)
		require.Equal(t, rawUrl, got.String())

		rawUrl = "file:/tmp/x"
		got, err = urlz.Parse(rawUrl)
		require.NoError(t, err)
		require.Equal(t, &urlz.Url{Scheme: "file", Path: p}, got)
		require.Equal(t, "file:///tmp/x", got.String())
	})
}
