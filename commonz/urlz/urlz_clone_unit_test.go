package urlz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/stretchr/testify/require"
)

func TestUnitUrlClone(t *testing.T) {
	type testScenario struct {
		name           string
		originalUrl    *urlz.Url
		expectedScheme string
		expectedPath   string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		clonedUrl := s.originalUrl.Clone()

		require.NotNil(t, clonedUrl)
		require.NotSame(t, s.originalUrl, clonedUrl, "Cloned URL should not be the same object as original")
		require.NotSame(t, s.originalUrl.Path, clonedUrl.Path, "Cloned URL's Path should not be the same object as original's Path")

		require.Equal(t, s.expectedScheme, clonedUrl.Scheme, "Scheme should match")
		require.Equal(t, s.originalUrl.User, clonedUrl.User, "User should match")
		require.Equal(t, s.originalUrl.Password, clonedUrl.Password, "Password should match")
		require.Equal(t, s.originalUrl.Host, clonedUrl.Host, "Host should match")
		require.Equal(t, s.originalUrl.Port, clonedUrl.Port, "Port should match")
		require.Equal(t, s.expectedPath, clonedUrl.Path.String(), "Path should match")
		require.Equal(t, s.originalUrl.Query, clonedUrl.Query, "Query should match")
		require.Equal(t, s.originalUrl.Fragment, clonedUrl.Fragment, "Fragment should match")

		// Ensure modifications to clone do not affect original
		clonedUrl.Scheme = "modified"
		clonedUrl.Path = pathz.MustParse("/modified/path")
		require.NotEqual(t, s.originalUrl.Scheme, clonedUrl.Scheme, "Modifying clone should not affect original scheme")
		require.NotEqual(t, s.originalUrl.Path.String(), clonedUrl.Path.String(), "Modifying clone should not affect original path")
	}

	t.Run("simple URL clone", func(t *testing.T) {
		originalPath := pathz.MustParse("/a/b/c")
		originalUrl := &urlz.Url{
			Scheme: "http",
			Host:   "example.com",
			Path:   originalPath,
			Query:  "param=value",
		}
		check(t, testScenario{
			name:           "simple URL clone",
			originalUrl:    originalUrl,
			expectedScheme: "http",
			expectedPath:   "/a/b/c",
		})
	})

	t.Run("URL with empty path clone", func(t *testing.T) {
		originalPath := pathz.MustParse("")
		originalUrl := &urlz.Url{
			Scheme: "file",
			Host:   "",
			Path:   originalPath,
		}
		check(t, testScenario{
			name:           "URL with empty path clone",
			originalUrl:    originalUrl,
			expectedScheme: "file",
			expectedPath:   "",
		})
	})

	t.Run("URL with root path clone", func(t *testing.T) {
		originalPath := pathz.MustParse("/")
		originalUrl := &urlz.Url{
			Scheme: "gs",
			Host:   "bucket",
			Path:   originalPath,
		}
		check(t, testScenario{
			name:           "URL with root path clone",
			originalUrl:    originalUrl,
			expectedScheme: "gs",
			expectedPath:   "/",
		})
	})
}
