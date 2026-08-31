package urlz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/stretchr/testify/require"
)

func TestUnitParseUnix(t *testing.T) {
	type testScenario struct {
		name        string
		rawUrl      string
		expected    *urlz.Url
		expectedStr string
		expectedErr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got, err := urlz.Parse(s.rawUrl)

		if s.expectedErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedErr)
			require.Nil(t, got)
			return
		}

		require.NoError(t, err)
		require.Equal(t, s.expected, got)
		if s.expectedStr != "" {
			require.Equal(t, s.expectedStr, got.String())
		}
	}

	p, err := pathz.Parse("/tmp/x.sock")
	require.NoError(t, err)

	t.Run("valid unix path", func(t *testing.T) {
		check(t, testScenario{
			name:   "valid unix path",
			rawUrl: "unix:/tmp/x.sock",
			expected: &urlz.Url{
				Scheme: "unix",
				Path:   p,
			},
			expectedStr: "unix:/tmp/x.sock",
		})
	})

	t.Run("valid unix path with opaque", func(t *testing.T) {
		check(t, testScenario{
			name:   "valid unix path with opaque",
			rawUrl: "unix:///tmp/x.sock",
			expected: &urlz.Url{
				Scheme: "unix",
				Path:   p,
			},
			expectedStr: "unix:/tmp/x.sock",
		})
	})

	t.Run("invalid relative path", func(t *testing.T) {
		check(t, testScenario{
			name:        "invalid relative path",
			rawUrl:      "unix:tmp/x.sock",
			expectedErr: "path absolute flag",
		})
	})

	t.Run("empty path", func(t *testing.T) {
		check(t, testScenario{
			name:        "empty path",
			rawUrl:      "unix:",
			expectedErr: "path absolute flag",
		})
	})
}
