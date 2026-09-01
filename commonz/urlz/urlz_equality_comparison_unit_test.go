package urlz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/stretchr/testify/require"
)

func TestUnitUrlEquality(t *testing.T) {
	type testScenario struct {
		name        string
		url1        string
		url2        string
		expectEqual bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		u1, err := urlz.Parse(s.url1)
		require.NoError(t, err)
		u2, err := urlz.Parse(s.url2)
		require.NoError(t, err)

		require.Equal(t, s.expectEqual, u1.Equals(u2), "Equality mismatch for %s", s.name)
		// Check for symmetry
		require.Equal(t, s.expectEqual, u2.Equals(u1), "Symmetry mismatch for %s", s.name)
	}

	t.Run("identical http urls", func(t *testing.T) {
		check(t, testScenario{
			name:        "identical http urls",
			url1:        "http://user:pass@example.com:8080/a/b?q=1#frag",
			url2:        "http://user:pass@example.com:8080/a/b?q=1#frag",
			expectEqual: true,
		})
	})

	t.Run("different schemes", func(t *testing.T) {
		check(t, testScenario{
			name:        "different schemes",
			url1:        "http://example.com/a",
			url2:        "https://example.com/a",
			expectEqual: false,
		})
	})

	t.Run("different hosts", func(t *testing.T) {
		check(t, testScenario{
			name:        "different hosts",
			url1:        "http://example.com/a",
			url2:        "http://example.org/a",
			expectEqual: false,
		})
	})

	t.Run("different paths", func(t *testing.T) {
		check(t, testScenario{
			name:        "different paths",
			url1:        "http://example.com/a",
			url2:        "http://example.com/b",
			expectEqual: false,
		})
	})

	t.Run("different users", func(t *testing.T) {
		check(t, testScenario{
			name:        "different users",
			url1:        "http://user1@example.com/a",
			url2:        "http://user2@example.com/a",
			expectEqual: false,
		})
	})

	t.Run("different passwords", func(t *testing.T) {
		check(t, testScenario{
			name:        "different passwords",
			url1:        "http://user:pass1@example.com/a",
			url2:        "http://user:pass2@example.com/a",
			expectEqual: false,
		})
	})

	t.Run("different ports", func(t *testing.T) {
		check(t, testScenario{
			name:        "different ports",
			url1:        "http://example.com:8080/a",
			url2:        "http://example.com:8081/a",
			expectEqual: false,
		})
	})

	t.Run("different queries", func(t *testing.T) {
		check(t, testScenario{
			name:        "different queries",
			url1:        "http://example.com/a?q=1",
			url2:        "http://example.com/a?q=2",
			expectEqual: false,
		})
	})

	t.Run("different fragments", func(t *testing.T) {
		check(t, testScenario{
			name:        "different fragments",
			url1:        "http://example.com/a#frag1",
			url2:        "http://example.com/a#frag2",
			expectEqual: false,
		})
	})

	t.Run("identical file urls", func(t *testing.T) {
		check(t, testScenario{
			name:        "identical file urls",
			url1:        "file:///a/b/c",
			url2:        "file:///a/b/c",
			expectEqual: true,
		})
	})

	t.Run("different file urls", func(t *testing.T) {
		check(t, testScenario{
			name:        "different file urls",
			url1:        "file:///a/b",
			url2:        "file:///a/c",
			expectEqual: false,
		})
	})

	t.Run("equivalent but different representation", func(t *testing.T) {
		u1, err := urlz.Parse("http://example.com/a/b")
		require.NoError(t, err)
		u2, err := urlz.Parse("http://example.com/a/../a/b")
		require.NoError(t, err)

		require.True(t, u1.Equals(u2))
	})

	t.Run("nil urls", func(t *testing.T) {
		var u1, u2 *urlz.Url
		require.True(t, u1.Equals(u2))

		u1, err := urlz.Parse("http://a.com")
		require.NoError(t, err)
		require.False(t, u1.Equals(u2))
		require.False(t, u2.Equals(u1))
	})
}

func TestUnitUrlComparison(t *testing.T) {
	type testScenario struct {
		name     string
		url1     string
		url2     string
		expected int
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		u1, err := urlz.Parse(s.url1)
		require.NoError(t, err)
		u2, err := urlz.Parse(s.url2)
		require.NoError(t, err)

		got := u1.Compare(u2)
		require.Equal(t, s.expected, got, "Comparison mismatch for %s: %s vs %s", s.name, s.url1, s.url2)
	}

	t.Run("identical urls", func(t *testing.T) {
		check(t, testScenario{
			name:     "identical urls",
			url1:     "http://example.com/a",
			url2:     "http://example.com/a",
			expected: 0,
		})
	})

	t.Run("different schemes", func(t *testing.T) {
		check(t, testScenario{
			name:     "different schemes",
			url1:     "http://example.com/a",
			url2:     "https://example.com/a",
			expected: -1, // http < https
		})
	})

	t.Run("different hosts", func(t *testing.T) {
		check(t, testScenario{
			name:     "different hosts",
			url1:     "http://a.example.com/a",
			url2:     "http://b.example.com/a",
			expected: -1, // a < b
		})
	})

	t.Run("different paths", func(t *testing.T) {
		check(t, testScenario{
			name:     "different paths",
			url1:     "http://example.com/a",
			url2:     "http://example.com/b",
			expected: -1, // /a < /b
		})
	})

	t.Run("different queries", func(t *testing.T) {
		check(t, testScenario{
			name:     "different queries",
			url1:     "http://example.com/a?q=1",
			url2:     "http://example.com/a?q=2",
			expected: -1, // q=1 < q=2
		})
	})

	t.Run("different fragments", func(t *testing.T) {
		check(t, testScenario{
			name:     "different fragments",
			url1:     "http://example.com/a#frag1",
			url2:     "http://example.com/a#frag2",
			expected: -1, // #frag1 < #frag2
		})
	})

	t.Run("url with and without query", func(t *testing.T) {
		check(t, testScenario{
			name:     "url with and without query",
			url1:     "http://example.com/a",
			url2:     "http://example.com/a?q=1",
			expected: -1,
		})
	})

	t.Run("url with and without fragment", func(t *testing.T) {
		check(t, testScenario{
			name:     "url with and without fragment",
			url1:     "http://example.com/a",
			url2:     "http://example.com/a#frag",
			expected: -1,
		})
	})

	t.Run("equivalent but different string representation", func(t *testing.T) {
		check(t, testScenario{
			name:     "equivalent but different string representation",
			url1:     "http://example.com/a/b",
			url2:     "http://example.com/a/../a/b",
			expected: 0, // /a/b and /a/../a/b are parsed to the same path
		})
	})
}
