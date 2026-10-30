package urlz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/stretchr/testify/require"
)

func TestUnitIs(t *testing.T) {
	type testScenario struct {
		name     string
		input    string
		expected bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got := urlz.Is(s.input)
		require.Equal(t, s.expected, got)
	}

	t.Run("http url returns true", func(t *testing.T) {
		check(t, testScenario{
			name:     "http url returns true",
			input:    "http://example.com",
			expected: true,
		})
	})

	t.Run("https url returns true", func(t *testing.T) {
		check(t, testScenario{
			name:     "https url returns true",
			input:    "https://example.com/api?foo=bar",
			expected: true,
		})
	})

	t.Run("custom scheme url returns true", func(t *testing.T) {
		check(t, testScenario{
			name:     "custom scheme url returns true",
			input:    "gs://bucket/object",
			expected: true,
		})
	})

	t.Run("scheme with digits returns true", func(t *testing.T) {
		check(t, testScenario{
			name:     "scheme with digits returns true",
			input:    "s3://bucket/key",
			expected: true,
		})
	})

	t.Run("unix socket url returns true", func(t *testing.T) {
		check(t, testScenario{
			name:     "unix socket url returns true",
			input:    "unix:/tmp/socket",
			expected: true,
		})
	})

	t.Run("absolute path returns false", func(t *testing.T) {
		check(t, testScenario{
			name:     "absolute path returns false",
			input:    "/users/123",
			expected: false,
		})
	})

	t.Run("relative path returns false", func(t *testing.T) {
		check(t, testScenario{
			name:     "relative path returns false",
			input:    "users/123",
			expected: false,
		})
	})

	t.Run("empty string returns false", func(t *testing.T) {
		check(t, testScenario{
			name:     "empty string returns false",
			input:    "",
			expected: false,
		})
	})

	t.Run("relative path with dots returns false", func(t *testing.T) {
		check(t, testScenario{
			name:     "relative path with dots returns false",
			input:    "../path/to/resource",
			expected: false,
		})
	})

	t.Run("scheme-relative url returns false", func(t *testing.T) {
		check(t, testScenario{
			name:     "scheme-relative url returns false",
			input:    "//example.com/api",
			expected: false,
		})
	})

	t.Run("port only returns false", func(t *testing.T) {
		check(t, testScenario{
			name:     "port only returns false",
			input:    ":8080",
			expected: false,
		})
	})
}
