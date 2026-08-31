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

	t.Run("http", func(t *testing.T) {
		p, err := pathz.Parse("/my/path")
		require.NoError(t, err)
		check(t, testScenario{
			name:   "full http url",
			rawUrl: "http://user:password@host.com:8080/my/path?query=1#fragment",
			expected: &urlz.Url{
				Scheme:   "http",
				User:     "user",
				Password: "password",
				Host:     "host.com",
				Port:     "8080",
				Path:     p,
				Query:    "query=1",
				Fragment: "fragment",
			},
		})
	})

	t.Run("https", func(t *testing.T) {
		p, err := pathz.Parse("/my/path")
		require.NoError(t, err)
		check(t, testScenario{
			name:   "full https url",
			rawUrl: "https://user:password@host.com:8080/my/path?query=1#fragment",
			expected: &urlz.Url{
				Scheme:   "https",
				User:     "user",
				Password: "password",
				Host:     "host.com",
				Port:     "8080",
				Path:     p,
				Query:    "query=1",
				Fragment: "fragment",
			},
		})
	})

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

	t.Run("gs", func(t *testing.T) {
		p, err := pathz.Parse("/my-folder/my-file")
		require.NoError(t, err)
		check(t, testScenario{
			name:   "google storage url",
			rawUrl: "gs://my-bucket/my-folder/my-file",
			expected: &urlz.Url{
				Scheme: "gs",
				Host:   "my-bucket",
				Path:   p,
			},
		})
	})

	t.Run("unix", func(t *testing.T) {
		p, err := pathz.Parse("/tmp/socket")
		require.NoError(t, err)

		rawUrl := "unix:/tmp/socket"
		got, err := urlz.Parse(rawUrl)
		require.NoError(t, err)
		require.Equal(t, &urlz.Url{Scheme: "unix", Path: p}, got)
		require.Equal(t, "unix:/tmp/socket", got.String())
	})

	t.Run("unknown scheme", func(t *testing.T) {
		check(t, testScenario{
			name:   "unknown scheme",
			rawUrl: "invalid-scheme://foo/bar",
			err:    "unknown scheme",
		})
	})
}
