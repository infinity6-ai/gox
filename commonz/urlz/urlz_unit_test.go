package urlz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/stretchr/testify/require"
)

func TestUnitParse(t *testing.T) {
	type testScenario struct {
		name     string
		rawUrl   string
		expected *urlz.Url
		err      string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got, err := urlz.Parse(s.rawUrl)
		if s.err != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.err)
			return
		}
		require.NoError(t, err)
		require.Equal(t, s.expected, got)

		// Also test String()
		if s.expected != nil {
			require.Equal(t, s.rawUrl, got.String())
		}
	}

	t.Run("unknown scheme", func(t *testing.T) {
		check(t, testScenario{
			name:   "unknown scheme",
			rawUrl: "invalid-scheme://foo/bar",
			err:    "unknown scheme",
		})
	})
}

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

func TestUnitUrlAppend(t *testing.T) {
	type testScenario struct {
		name        string
		originalUrl *urlz.Url
		appendPaths []*pathz.Path
		expectedUrl string
		expectedErr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		originalPathString := s.originalUrl.Path.String() // Save original path string for immutability check
		appendedUrl, err := s.originalUrl.Append(s.appendPaths...)

		if s.expectedErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedErr)
			require.Nil(t, appendedUrl)
			// Ensure original URL is unchanged on error
			require.Equal(t, originalPathString, s.originalUrl.Path.String(), "Original URL's path should be unchanged on error")
			return
		}

		require.NoError(t, err)
		require.NotNil(t, appendedUrl)
		require.NotSame(t, s.originalUrl, appendedUrl, "Appended URL should be a new object")
		require.NotSame(t, s.originalUrl.Path, appendedUrl.Path, "Appended URL's Path should be a new object")

		// Verify the new URL's string representation
		require.Equal(t, s.expectedUrl, appendedUrl.String(), "Appended URL string should match expected")

		// Ensure original URL is unchanged
		require.Equal(t, originalPathString, s.originalUrl.Path.String(), "Original URL's path should be unchanged after append")
	}

	t.Run("append single relative path", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "http", Host: "example.com", Path: pathz.MustParse("/base")}
		check(t, testScenario{
			name:        "append single relative path",
			originalUrl: originalUrl,
			appendPaths: []*pathz.Path{pathz.MustParse("segment")},
			expectedUrl: "http://example.com/base/segment",
		})
	})

	t.Run("append multiple relative paths", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "http", Host: "example.com", Path: pathz.MustParse("/base/folder/")}
		check(t, testScenario{
			name:        "append multiple relative paths",
			originalUrl: originalUrl,
			appendPaths: []*pathz.Path{pathz.MustParse("sub"), pathz.MustParse("file.txt")},
			expectedUrl: "http://example.com/base/folder/sub/file.txt",
		})
	})

	t.Run("append with parent traversal", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "file", Path: pathz.MustParse("/usr/local/bin")}
		check(t, testScenario{
			name:        "append with parent traversal",
			originalUrl: originalUrl,
			appendPaths: []*pathz.Path{pathz.MustParse("../share")},
			expectedErr: "path escaped error: joining '/usr/local/bin' to '[../share]' results in '/usr/local/share' which is outside the base: error appending path to url file:///usr/local/bin: [../share]",
		})
	})

	t.Run("append resulting in escaped path", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "http", Host: "example.com", Path: pathz.MustParse("/a/b")}
		check(t, testScenario{
			name:        "append resulting in escaped path",
			originalUrl: originalUrl,
			appendPaths: []*pathz.Path{pathz.MustParse("../../c")},
			expectedErr: "path escaped error: joining '/a/b' to '[../../c]' results in '/c' which is outside the base: error appending path to url http://example.com/a/b: [../../c]",
		})
	})

	t.Run("append empty path", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "http", Host: "example.com", Path: pathz.MustParse("/base")}
		check(t, testScenario{
			name:        "append empty path",
			originalUrl: originalUrl,
			appendPaths: []*pathz.Path{pathz.MustParse("")},
			expectedUrl: "http://example.com/base",
		})
	})

	t.Run("append to root path", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "file", Path: pathz.MustParse("/")}
		check(t, testScenario{
			name:        "append to root path",
			originalUrl: originalUrl,
			appendPaths: []*pathz.Path{pathz.MustParse("folder/file.txt")},
			expectedUrl: "file:///folder/file.txt",
		})
	})

	t.Run("append absolute path should discard previous and return escaped error", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "http", Host: "example.com", Path: pathz.MustParse("/base")}
		check(t, testScenario{
			name:        "append absolute path should discard previous and return escaped error",
			originalUrl: originalUrl,
			appendPaths: []*pathz.Path{pathz.MustParse("/new/absolute/path")},
			expectedErr: "path escaped error: joining '/base' to '[/new/absolute/path]' results in '/new/absolute/path' which is outside the base: error appending path to url http://example.com/base: [/new/absolute/path]",
		})
	})

	t.Run("append with query parameters (should not affect query)", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "http", Host: "example.com", Path: pathz.MustParse("/base"), Query: "q=test"}
		check(t, testScenario{
			name:        "append with query parameters",
			originalUrl: originalUrl,
			appendPaths: []*pathz.Path{pathz.MustParse("segment")},
			expectedUrl: "http://example.com/base/segment?q=test",
		})
	})

	t.Run("append with fragment (should not affect fragment)", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "http", Host: "example.com", Path: pathz.MustParse("/base"), Fragment: "anchor"}
		check(t, testScenario{
			name:        "append with fragment",
			originalUrl: originalUrl,
			appendPaths: []*pathz.Path{pathz.MustParse("segment")},
			expectedUrl: "http://example.com/base/segment#anchor",
		})
	})
}
