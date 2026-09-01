package urlz_test

import (
	"testing"

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
