package urlz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitSimpleUrl(t *testing.T) {
	type testScenario struct {
		name      string
		url       string
		expectErr string
		check     func(t *testing.T, u *SimpleUrl)
	}

	scenarios := []testScenario{
		{
			name: "valid file url",
			url:  "file:///path/to/file",
			check: func(t *testing.T, u *SimpleUrl) {
				require.Equal(t, "file", u.Schema())
				require.Equal(t, "///path/to/file", u.path)
				require.Equal(t, "file:///path/to/file", u.String())
			},
		},
		{
			name: "valid gs url",
			url:  "gs://bucket/object",
			check: func(t *testing.T, u *SimpleUrl) {
				require.Equal(t, "gs", u.Schema())
				require.Equal(t, "//bucket/object", u.path)
				require.Equal(t, "gs://bucket/object", u.String())
			},
		},
		{
			name: "valid unix url",
			url:  "unix:/tmp/socket",
			check: func(t *testing.T, u *SimpleUrl) {
				require.Equal(t, "unix", u.Schema())
				require.Equal(t, "/tmp/socket", u.path)
				require.Equal(t, "unix:/tmp/socket", u.String())
			},
		},
		{
			name:      "invalid schema",
			url:       "http://google.com",
			expectErr: "invalid schema for SimpleUrl: http",
		},
		{
			name:      "no schema",
			url:       "just/a/path",
			expectErr: "unknown schema",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			u, err := ParseSimpleUrl(s.url)

			if s.expectErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), s.expectErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, u)
			if s.check != nil {
				s.check(t, u)
			}
		})
	}
}
