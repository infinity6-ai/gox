package urlz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/stretchr/testify/require"
)

func TestUnitParseUnix(t *testing.T) {
	t.Run("unix", func(t *testing.T) {
		p, err := pathz.Parse("/tmp/socket")
		require.NoError(t, err)

		rawUrl := "unix:/tmp/socket"
		got, err := urlz.Parse(rawUrl)
		require.NoError(t, err)
		require.Equal(t, &urlz.Url{Scheme: "unix", Path: p}, got)
		require.Equal(t, "unix:/tmp/socket", got.String())
	})
}
