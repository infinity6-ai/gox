package urlz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/stretchr/testify/require"
)

func TestUnitUrlIsBaseOf(t *testing.T) {
	type testScenario struct {
		name     string
		baseUrl  string
		otherUrl string
		isBase   bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		base, err := urlz.Parse(s.baseUrl)
		require.NoError(t, err, "parsing baseUrl failed")
		other, err := urlz.Parse(s.otherUrl)
		require.NoError(t, err, "parsing otherUrl failed")

		require.Equal(t, s.isBase, base.IsBaseOf(other))
	}

	t.Run("identical http urls", func(t *testing.T) {
		check(t, testScenario{
			name:     "identical http urls",
			baseUrl:  "http://example.com/a/b",
			otherUrl: "http://example.com/a/b",
			isBase:   true,
		})
	})

	t.Run("simple base path", func(t *testing.T) {
		check(t, testScenario{
			name:     "simple base path",
			baseUrl:  "http://example.com/a",
			otherUrl: "http://example.com/a/b",
			isBase:   true,
		})
	})

	t.Run("with trailing slash on base", func(t *testing.T) {
		check(t, testScenario{
			name:     "with trailing slash on base",
			baseUrl:  "http://example.com/a/",
			otherUrl: "http://example.com/a/b",
			isBase:   true,
		})
	})

	t.Run("with trailing slash on both", func(t *testing.T) {
		check(t, testScenario{
			name:     "with trailing slash on both",
			baseUrl:  "http://example.com/a/",
			otherUrl: "http://example.com/a/b/",
			isBase:   true,
		})
	})

	t.Run("ignores query params on other", func(t *testing.T) {
		check(t, testScenario{
			name:     "ignores query params on other",
			baseUrl:  "http://example.com/a",
			otherUrl: "http://example.com/a/b?q=1",
			isBase:   true,
		})
	})

	t.Run("ignores fragment on other", func(t *testing.T) {
		check(t, testScenario{
			name:     "ignores fragment on other",
			baseUrl:  "http://example.com/a",
			otherUrl: "http://example.com/a/b#frag",
			isBase:   true,
		})
	})

	t.Run("different scheme", func(t *testing.T) {
		check(t, testScenario{
			name:     "different scheme",
			baseUrl:  "http://example.com/a",
			otherUrl: "https://example.com/a/b",
			isBase:   false,
		})
	})

	t.Run("different host", func(t *testing.T) {
		check(t, testScenario{
			name:     "different host",
			baseUrl:  "http://example.com/a",
			otherUrl: "http://another.com/a/b",
			isBase:   false,
		})
	})

	t.Run("different path", func(t *testing.T) {
		check(t, testScenario{
			name:     "different path",
			baseUrl:  "http://example.com/a",
			otherUrl: "http://example.com/c/b",
			isBase:   false,
		})
	})

	t.Run("not a base path", func(t *testing.T) {
		check(t, testScenario{
			name:     "not a base path",
			baseUrl:  "http://example.com/a/b",
			otherUrl: "http://example.com/a",
			isBase:   false,
		})
	})

	t.Run("different user", func(t *testing.T) {
		check(t, testScenario{
			name:     "different user",
			baseUrl:  "http://user1@example.com/a",
			otherUrl: "http://user2@example.com/a",
			isBase:   false,
		})
	})

	t.Run("file scheme", func(t *testing.T) {
		check(t, testScenario{
			name:     "file scheme",
			baseUrl:  "file:///a/b",
			otherUrl: "file:///a/b/c/d",
			isBase:   true,
		})
	})

	t.Run("file scheme different path", func(t *testing.T) {
		check(t, testScenario{
			name:     "file scheme different path",
			baseUrl:  "file:///a/c",
			otherUrl: "file:///a/b/c",
			isBase:   false,
		})
	})

	t.Run("identical with credentials", func(t *testing.T) {
		check(t, testScenario{
			name:     "identical with credentials",
			baseUrl:  "http://user:pass@example.com/a",
			otherUrl: "http://user:pass@example.com/a/b",
			isBase:   true,
		})
	})

	t.Run("different credentials", func(t *testing.T) {
		check(t, testScenario{
			name:     "different credentials",
			baseUrl:  "http://user:pass1@example.com/a",
			otherUrl: "http://user:pass2@example.com/a/b",
			isBase:   false,
		})
	})
}
