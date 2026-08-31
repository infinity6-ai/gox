package urlz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/stretchr/testify/require"
)

func TestUnitParseGs(t *testing.T) {
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
}
