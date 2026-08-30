package urlz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitHttpUrl(t *testing.T) {
	type testScenario struct {
		name      string
		schema    string
		urlPart   string
		expectErr string
		expectURL string
		check     func(t *testing.T, u *HttpUrl)
	}

	scenarios := []testScenario{
		{
			name:    "valid https",
			schema:  "https",
			urlPart: "//google.com/search?q=test",
			expectURL: "https://google.com/search?q=test",
			check: func(t *testing.T, u *HttpUrl) {
				require.Equal(t, "https", u.Schema())
				require.Equal(t, "google.com", u.host)
				require.Equal(t, "", u.port)
				require.Equal(t, "/search", u.path)
				require.Equal(t, "q=test", u.query)
				require.Equal(t, "", u.fragment)
			},
		},
		{
			name:    "valid http with port",
			schema:  "http",
			urlPart: "localhost:8080/path",
			expectURL: "http://localhost:8080/path",
			check: func(t *testing.T, u *HttpUrl) {
				require.Equal(t, "http", u.Schema())
				require.Equal(t, "localhost", u.host)
				require.Equal(t, "8080", u.port)
				require.Equal(t, "/path", u.path)
			},
		},
		{
			name:    "url without slashes",
			schema:  "https",
			urlPart: "example.com",
			expectURL: "https://example.com",
			check: func(t *testing.T, u *HttpUrl) {
				require.Equal(t, "https", u.Schema())
				require.Equal(t, "example.com", u.host)
				require.Equal(t, "", u.path)
			},
		},
		{
			name:      "invalid schema",
			schema:    "ftp",
			urlPart:   "//example.com",
			expectErr: "invalid schema for HttpUrl: ftp",
		},
		{
			name:      "no host",
			schema:    "http",
			urlPart:   "/path/only",
			expectErr: "host cannot be empty for HttpUrl",
		},
		{
			name:    "with user info",
			schema:  "https",
			urlPart: "//user:pass@example.com",
			expectURL: "https://user:pass@example.com",
			check: func(t *testing.T, u *HttpUrl) {
				require.NotNil(t, u.userInfo)
				require.Equal(t, "user", u.userInfo.Username())
				p, ok := u.userInfo.Password()
				require.True(t, ok)
				require.Equal(t, "pass", p)
			},
		},
		{
			name:    "full url",
			schema:  "https",
			urlPart: "//user:password@example.com:8080/path/to/resource?query=value#fragment",
			expectURL: "https://user:password@example.com:8080/path/to/resource?query=value#fragment",
			check: func(t *testing.T, u *HttpUrl) {
				require.Equal(t, "https", u.schema)
				require.Equal(t, "example.com", u.host)
				require.Equal(t, "8080", u.port)
				require.Equal(t, "/path/to/resource", u.path)
				require.Equal(t, "query=value", u.query)
				require.Equal(t, "fragment", u.fragment)
				require.NotNil(t, u.userInfo)
				require.Equal(t, "user", u.userInfo.Username())
				p, ok := u.userInfo.Password()
				require.True(t, ok)
				require.Equal(t, "password", p)
			},
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			u, err := NewHttpUrl(s.schema, s.urlPart)

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
			require.Equal(t, s.expectURL, u.String())
		})
	}
}
