package urlz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/stretchr/testify/require"
)

func TestUnitUrlJoinPath(t *testing.T) {
	type testScenario struct {
		name string
		originalUrl *urlz.Url
		joinPaths []*pathz.Path
		expectedUrl string
		expectedErr string
		expectedUrlOnError *urlz.Url // New field for error-case URL assertion
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		originalPathString := s.originalUrl.Path.String() // Save original path string for immutability check
		joinedUrl, err := s.originalUrl.JoinPath(s.joinPaths...)

		if s.expectedErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedErr)
			require.NotNil(t, joinedUrl) // joinedUrl should not be nil even on error
			require.Equal(t, s.expectedUrlOnError, joinedUrl, "Joined URL on error should match expected error URL")
			// Ensure original URL is unchanged on error
			require.Equal(t, originalPathString, s.originalUrl.Path.String(), "Original URL's path should be unchanged on error")
			return
		}

		require.NoError(t, err)
		require.NotNil(t, joinedUrl)
		require.NotSame(t, s.originalUrl, joinedUrl, "Joined URL should be a new object")
		require.NotSame(t, s.originalUrl.Path, joinedUrl.Path, "Joined URL's Path should be a new object")

		// Verify the new URL's string representation
		require.Equal(t, s.expectedUrl, joinedUrl.String(), "Joined URL string should match expected")

		// Ensure original URL is unchanged
		require.Equal(t, originalPathString, s.originalUrl.Path.String(), "Original URL's path should be unchanged after join")
	}

	t.Run("join single relative path", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "http", Host: "example.com", Path: pathz.MustParse("/base")}
		check(t, testScenario{
			name:        "join single relative path",
			originalUrl: originalUrl,
			joinPaths: []*pathz.Path{pathz.MustParse("segment")},
			expectedUrl: "http://example.com/base/segment",
		})
	})

	t.Run("join multiple relative paths", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "http", Host: "example.com", Path: pathz.MustParse("/base/folder/")}
		check(t, testScenario{
			name:        "join multiple relative paths",
			originalUrl: originalUrl,
			joinPaths: []*pathz.Path{pathz.MustParse("sub"), pathz.MustParse("file.txt")},
			expectedUrl: "http://example.com/base/folder/sub/file.txt",
		})
	})

	t.Run("join with parent traversal", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "file", Path: pathz.MustParse("/usr/local/bin")}
		check(t, testScenario{
			name:        "join with parent traversal",
			originalUrl: originalUrl,
			joinPaths: []*pathz.Path{pathz.MustParse("../share")},
			expectedErr: "path escaped error: joining '/usr/local/bin' to '[../share]' results in '/usr/local/share' which is outside the base: error joining path to url file:///usr/local/bin: [../share]",
			expectedUrlOnError: &urlz.Url{
				Scheme: "file",
				Path:   pathz.MustParse("/usr/local/share"),
			},
		})
	})

	t.Run("join with boxlocal scheme", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "boxlocal", Path: pathz.MustParse("/my/box")}
		check(t, testScenario{
			name:        "join with boxlocal scheme",
			originalUrl: originalUrl,
			joinPaths:   []*pathz.Path{pathz.MustParse("item")},
			expectedUrl: "boxlocal:///my/box/item",
		})
	})

	t.Run("join resulting in escaped path", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "http", Host: "example.com", Path: pathz.MustParse("/a/b")}
		check(t, testScenario{
			name:        "join resulting in escaped path",
			originalUrl: originalUrl,
			joinPaths: []*pathz.Path{pathz.MustParse("../../c")},
			expectedErr: "path escaped error: joining '/a/b' to '[../../c]' results in '/c' which is outside the base: error joining path to url http://example.com/a/b: [../../c]",
			expectedUrlOnError: &urlz.Url{
				Scheme: "http",
				Host:   "example.com",
				Path:   pathz.MustParse("/c"),
			},
		})
	})

	t.Run("join empty path", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "http", Host: "example.com", Path: pathz.MustParse("/base")}
		check(t, testScenario{
			name:        "join empty path",
			originalUrl: originalUrl,
			joinPaths: []*pathz.Path{pathz.MustParse("")},
			expectedUrl: "http://example.com/base",
		})
	})

	t.Run("join to root path", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "file", Path: pathz.MustParse("/")}
		check(t, testScenario{
			name:        "join to root path",
			originalUrl: originalUrl,
			joinPaths: []*pathz.Path{pathz.MustParse("folder/file.txt")},
			expectedUrl: "file:///folder/file.txt",
		})
	})

	t.Run("join absolute path should discard previous and return escaped error", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "http", Host: "example.com", Path: pathz.MustParse("/base")}
		check(t, testScenario{
			name:        "join absolute path should discard previous and return escaped error",
			originalUrl: originalUrl,
			joinPaths: []*pathz.Path{pathz.MustParse("/new/absolute/path")},
			expectedErr: "path escaped error: joining '/base' to '[/new/absolute/path]' results in '/new/absolute/path' which is outside the base: error joining path to url http://example.com/base: [/new/absolute/path]",
			expectedUrlOnError: &urlz.Url{
				Scheme: "http",
				Host:   "example.com",
				Path:   pathz.MustParse("/new/absolute/path"),
			},
		})
	})

	t.Run("join with query parameters (should not affect query)", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "http", Host: "example.com", Path: pathz.MustParse("/base"), Query: "q=test"}
		check(t, testScenario{
			name:        "join with query parameters",
			originalUrl: originalUrl,
			joinPaths: []*pathz.Path{pathz.MustParse("segment")},
			expectedUrl: "http://example.com/base/segment?q=test",
		})
	})

	t.Run("join with fragment (should not affect fragment)", func(t *testing.T) {
		originalUrl := &urlz.Url{Scheme: "http", Host: "example.com", Path: pathz.MustParse("/base"), Fragment: "anchor"}
		check(t, testScenario{
			name:        "join with fragment",
			originalUrl: originalUrl,
			joinPaths: []*pathz.Path{pathz.MustParse("segment")},
			expectedUrl: "http://example.com/base/segment#anchor",
		})
	})
}
