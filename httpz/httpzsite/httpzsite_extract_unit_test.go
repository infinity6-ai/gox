package httpzsite_test

import (
	"errors"
	"os"
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/trokerz"
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

		err := site.ExtractTo(ctx, httpzsite.ExtractOptions{
			Path: dest,
		})
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

	t.Run("Extract with nil path returns error", func(t *testing.T) {
		site := &httpzsite.Site{
			Context: ctx,
			Pack:    testPackName,
		}

		err := site.ExtractTo(ctx, httpzsite.ExtractOptions{
			Path: nil,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "path is required")
	})

	t.Run("Extract invalid pack returns error", func(t *testing.T) {
		site := &httpzsite.Site{
			Context: ctx,
			Pack:    "nonexistent-pack",
		}

		tempDir := t.TempDir()
		dest := pathz.MustParse(tempDir)

		err := site.ExtractTo(ctx, httpzsite.ExtractOptions{
			Path: dest,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to extract site pack")
	})

	t.Run("Extract with troke enabled replaces placeholders and preserves binary files", func(t *testing.T) {
		site := &httpzsite.Site{
			Context: ctx,
			Pack:    testTemplatePackName,
			TrokerOptions: []trokerz.Options{
				{
					Begin: "{{",
					End:   "}}",
					Get: func(k string) (string, bool, error) {
						switch k {
						case "title":
							return "My App", true, nil
						case "API_URL":
							return "https://example.com/api", true, nil
						default:
							return "", false, nil
						}
					},
				},
			},
		}

		tempDir := t.TempDir()
		dest := pathz.MustParse(tempDir)

		err := site.ExtractTo(ctx, httpzsite.ExtractOptions{
			Path:  dest,
			Troke: true,
		})
		require.NoError(t, err)

		indexFile := dest.MustJoin(pathz.MustParse("index.html"))
		content, err := os.ReadFile(indexFile.String())
		require.NoError(t, err)
		require.Equal(t, "<!DOCTYPE html><html><body>My App</body></html>", string(content))

		configFile := dest.MustJoin(pathz.MustParse("config.js"))
		content, err = os.ReadFile(configFile.String())
		require.NoError(t, err)
		require.Equal(t, "const api = 'https://example.com/api';", string(content))

		binFile := dest.MustJoin(pathz.MustParse("raw.bin"))
		binContent, err := os.ReadFile(binFile.String())
		require.NoError(t, err)
		require.Equal(t, []byte{0xff, 0xfe, 0xfd, 0x00}, binContent)
	})

	t.Run("Extract with troke disabled keeps placeholders unreplaced", func(t *testing.T) {
		site := &httpzsite.Site{
			Context: ctx,
			Pack:    testTemplatePackName,
			TrokerOptions: []trokerz.Options{
				{
					Begin: "{{",
					End:   "}}",
					Get: func(k string) (string, bool, error) {
						if k == "title" {
							return "My App", true, nil
						}
						return "", false, nil
					},
				},
			},
		}

		tempDir := t.TempDir()
		dest := pathz.MustParse(tempDir)

		err := site.ExtractTo(ctx, httpzsite.ExtractOptions{
			Path:  dest,
			Troke: false,
		})
		require.NoError(t, err)

		indexFile := dest.MustJoin(pathz.MustParse("index.html"))
		content, err := os.ReadFile(indexFile.String())
		require.NoError(t, err)
		require.Equal(t, "<!DOCTYPE html><html><body>{{title}}</body></html>", string(content))
	})

	t.Run("Extract with troke error returns error", func(t *testing.T) {
		site := &httpzsite.Site{
			Context: ctx,
			Pack:    testTemplatePackName,
			TrokerOptions: []trokerz.Options{
				{
					Begin: "{{",
					End:   "}}",
					Get: func(k string) (string, bool, error) {
						return "", false, errors.New("lookup failed")
					},
				},
			},
		}

		tempDir := t.TempDir()
		dest := pathz.MustParse(tempDir)

		err := site.ExtractTo(ctx, httpzsite.ExtractOptions{
			Path:  dest,
			Troke: true,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to extract site pack")
	})
}
