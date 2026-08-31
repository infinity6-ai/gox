package urlz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/stretchr/testify/require"
)

func TestUnitParseFile(t *testing.T) {
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

	p, err := pathz.Parse("/tmp/x")
	require.NoError(t, err)

	t.Run("absolute path with slashes", func(t *testing.T) {
		check(t, testScenario{
			name:   "absolute path with slashes",
			rawUrl: "file:///tmp/x",
			expected: &urlz.Url{
				Scheme: "file",
				Path:   p,
			},
			expectedStr: "file:///tmp/x",
		})
	})

	t.Run("absolute path", func(t *testing.T) {
		check(t, testScenario{
			name:   "absolute path",
			rawUrl: "file:/tmp/x",
			expected: &urlz.Url{
				Scheme: "file",
				Path:   p,
			},
			expectedStr: "file:///tmp/x",
		})
	})

	t.Run("relative path", func(t *testing.T) {
		check(t, testScenario{
			name:        "relative path",
			rawUrl:      "file:tmp/x",
			expectedErr: "path absolute flag",
		})
	})
}
