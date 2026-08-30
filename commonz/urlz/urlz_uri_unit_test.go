package urlz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitParse(t *testing.T) {
	type testScenario struct {
		name      string
		url       string
		expectErr string
		check     func(t *testing.T, u URI)
	}

	scenarios := []testScenario{
		{
			name: "parse http",
			url:  "http://google.com",
			check: func(t *testing.T, u URI) {
				require.IsType(t, &HttpUrl{}, u)
				require.Equal(t, "http", u.Schema())
				require.Equal(t, "http://google.com", u.String())
			},
		},
		{
			name: "parse https",
			url:  "https://google.com",
			check: func(t *testing.T, u URI) {
				require.IsType(t, &HttpUrl{}, u)
				require.Equal(t, "https", u.Schema())
				require.Equal(t, "https://google.com", u.String())
			},
		},
		{
			name: "parse file",
			url:  "file:///path/to/file",
			check: func(t *testing.T, u URI) {
				require.IsType(t, &SimpleUrl{}, u)
				require.Equal(t, "file", u.Schema())
				require.Equal(t, "file:///path/to/file", u.String())
			},
		},
		{
			name: "parse gs",
			url:  "gs://bucket/object",
			check: func(t *testing.T, u URI) {
				require.IsType(t, &SimpleUrl{}, u)
				require.Equal(t, "gs", u.Schema())
				require.Equal(t, "gs://bucket/object", u.String())
			},
		},
		{
			name: "parse unix",
			url:  "unix:/tmp/socket",
			check: func(t *testing.T, u URI) {
				require.IsType(t, &SimpleUrl{}, u)
				require.Equal(t, "unix", u.Schema())
				require.Equal(t, "unix:/tmp/socket", u.String())
			},
		},
		{
			name:      "unknown schema",
			url:       "foo:bar",
			expectErr: "unknown schema: foo",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			u, err := Parse(s.url)
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
