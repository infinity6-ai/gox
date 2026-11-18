package httpzsite_test

import (
	"os"
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/httpz/httpzsite"
	"github.com/stretchr/testify/require"
)

func TestUnitSiteExtractTo(t *testing.T) {
	ctx := t.Context()

	t.Run("Extract all pack files to destination directory", func(t *testing.T) {
		site := &httpzsite.Site{
			Context: ctx,
			Pack:    testPackName,
		}

		tempDir := t.TempDir()
		dest := pathz.MustParse(tempDir)

		err := site.ExtractTo(ctx, dest)
		require.NoError(t, err)

		expectedFiles := map[string]string{
			"index.html":        "<html><body>Home</body></html>",
			"style.css":         "body { color: red; }",
			"sub/index.html":    "<html><body>Sub</body></html>",
			"sub/script.js":     "console.log('sub');",
			"deep/nested/a.txt": "deep-content",
		}

		for relPath, expectedContent := range expectedFiles {
			p := dest.MustJoin(pathz.MustParse(relPath))
			content, err := os.ReadFile(p.String())
			require.NoError(t, err, "failed to read extracted file: %s", p.String())
			require.Equal(t, expectedContent, string(content))
		}
	})

	t.Run("Extract invalid pack returns error", func(t *testing.T) {
		site := &httpzsite.Site{
			Context: ctx,
			Pack:    "nonexistent-pack",
		}

		tempDir := t.TempDir()
		dest := pathz.MustParse(tempDir)

		err := site.ExtractTo(ctx, dest)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to extract site pack")
	})
}
