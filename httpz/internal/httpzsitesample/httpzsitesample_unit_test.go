package httpzsitesample_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/httpz/httpzsite"
	"github.com/infinity6-ai/gox/httpz/internal/httpzsitesample"
	"github.com/stretchr/testify/require"
)

func TestUnitSampleSite(t *testing.T) {
	site := httpzsitesample.New()
	require.NotNil(t, site)
	require.Equal(t, "localhost:8080", site.LocalAddress)
	require.Equal(t, httpzsitesample.PackName, site.Pack)

	t.Run("Extract sample site pack", func(t *testing.T) {
		tempDir := t.TempDir()
		err := site.ExtractTo(t.Context(), httpzsite.ExtractOptions{
			Path: pathz.MustParse(tempDir),
		})
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tempDir, "index.html"))
		require.NoError(t, err)
		require.Contains(t, string(content), "Hello from Sample Site!")

		css, err := os.ReadFile(filepath.Join(tempDir, "style.css"))
		require.NoError(t, err)
		require.Contains(t, string(css), "font-family")
	})
}
