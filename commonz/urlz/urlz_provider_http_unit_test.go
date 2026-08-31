package urlz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/stretchr/testify/require"
)

func TestUnitParseHttp(t *testing.T) {
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

		check(t, testScenario{
			name:   "http no user/password",
			rawUrl: "http://host.com:8080/my/path?query=1#fragment",
			expected: &urlz.Url{
				Scheme:   "http",
				Host:     "host.com",
				Port:     "8080",
				Path:     p,
				Query:    "query=1",
				Fragment: "fragment",
			},
		})

		p2, err := pathz.Parse("/my/path")
		require.NoError(t, err)
		check(t, testScenario{
			name:   "http no port",
			rawUrl: "http://user:password@host.com/my/path?query=1#fragment",
			expected: &urlz.Url{
				Scheme:   "http",
				User:     "user",
				Password: "password",
				Host:     "host.com",
				Path:     p2,
				Query:    "query=1",
				Fragment: "fragment",
			},
		})

		check(t, testScenario{
			name:   "http no path",
			rawUrl: "http://user:password@host.com:8080?query=1#fragment",
			expected: &urlz.Url{
				Scheme:   "http",
				User:     "user",
				Password: "password",
				Host:     "host.com",
				Port:     "8080",
				Path:     pathz.New(0, nil, false),
				Query:    "query=1",
				Fragment: "fragment",
			},
		})

		check(t, testScenario{
			name:   "http no query",
			rawUrl: "http://user:password@host.com:8080/my/path#fragment",
			expected: &urlz.Url{
				Scheme:   "http",
				User:     "user",
				Password: "password",
				Host:     "host.com",
				Port:     "8080",
				Path:     p,
				Fragment: "fragment",
			},
		})

		check(t, testScenario{
			name:   "http no fragment",
			rawUrl: "http://user:password@host.com:8080/my/path?query=1",
			expected: &urlz.Url{
				Scheme:   "http",
				User:     "user",
				Password: "password",
				Host:     "host.com",
				Port:     "8080",
				Path:     p,
				Query:    "query=1",
			},
		})

		check(t, testScenario{
			name:   "http host only",
			rawUrl: "http://host.com",
			expected: &urlz.Url{
				Scheme: "http",
				Host:   "host.com",
				Path:   pathz.New(0, nil, false),
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

		check(t, testScenario{
			name:   "https no user/password",
			rawUrl: "https://host.com:8080/my/path?query=1#fragment",
			expected: &urlz.Url{
				Scheme:   "https",
				Host:     "host.com",
				Port:     "8080",
				Path:     p,
				Query:    "query=1",
				Fragment: "fragment",
			},
		})

		p2, err := pathz.Parse("/my/path")
		require.NoError(t, err)
		check(t, testScenario{
			name:   "https no port",
			rawUrl: "https://user:password@host.com/my/path?query=1#fragment",
			expected: &urlz.Url{
				Scheme:   "https",
				User:     "user",
				Password: "password",
				Host:     "host.com",
				Path:     p2,
				Query:    "query=1",
				Fragment: "fragment",
			},
		})

		check(t, testScenario{
			name:   "https no path",
			rawUrl: "https://user:password@host.com:8080?query=1#fragment",
			expected: &urlz.Url{
				Scheme:   "https",
				User:     "user",
				Password: "password",
				Host:     "host.com",
				Port:     "8080",
				Path:     pathz.New(0, nil, false),
				Query:    "query=1",
				Fragment: "fragment",
			},
		})

		check(t, testScenario{
			name:   "https no query",
			rawUrl: "https://user:password@host.com:8080/my/path#fragment",
			expected: &urlz.Url{
				Scheme:   "https",
				User:     "user",
				Password: "password",
				Host:     "host.com",
				Port:     "8080",
				Path:     p,
				Fragment: "fragment",
			},
		})

		check(t, testScenario{
			name:   "https no fragment",
			rawUrl: "https://user:password@host.com:8080/my/path?query=1",
			expected: &urlz.Url{
				Scheme:   "https",
				User:     "user",
				Password: "password",
				Host:     "host.com",
				Port:     "8080",
				Path:     p,
				Query:    "query=1",
			},
		})

		check(t, testScenario{
			name:   "https host only",
			rawUrl: "https://host.com",
			expected: &urlz.Url{
				Scheme: "https",
				Host:   "host.com",
				Path:   pathz.New(0, nil, false),
			},
		})
	})

	t.Run("invalid urls", func(t *testing.T) {
		check(t, testScenario{
			name:   "invalid scheme",
			rawUrl: "ftp://host.com",
			err:    "unknown scheme",
		})
		check(t, testScenario{
			name:   "missing host",
			rawUrl: "http:///path",
			err:    "validation error must not be empty: host",
		})
		check(t, testScenario{
			name:   "invalid port",
			rawUrl: "http://host.com:abc",
			err:    "invalid port \":abc\" after host",
		})
		check(t, testScenario{
			name:   "invalid ip host",
			rawUrl: "http://192.168.1.1.1/path",
			err:    "invalid http url",
		})
		check(t, testScenario{
			name:   "invalid ipv6 host",
			rawUrl: "http://[::1::]/path",
			err:    "invalid http url",
		})
	})

	t.Run("ip hosts", func(t *testing.T) {
		p, err := pathz.Parse("/path")
		require.NoError(t, err)
		check(t, testScenario{
			name:   "ipv4 host",
			rawUrl: "http://192.168.1.1/path",
			expected: &urlz.Url{
				Scheme: "http",
				Host:   "192.168.1.1",
				Path:   p,
			},
		})
		check(t, testScenario{
			name:   "ipv6 host",
			rawUrl: "http://[::1]/path",
			expected: &urlz.Url{
				Scheme: "http",
				Host:   "::1",
				Path:   p,
			},
		})
		check(t, testScenario{
			name:   "ipv6 host with port",
			rawUrl: "http://[::1]:8080/path",
			expected: &urlz.Url{
				Scheme: "http",
				Host:   "::1",
				Port:   "8080",
				Path:   p,
			},
		})
	})
}
