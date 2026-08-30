package pathz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitExtractRelative(t *testing.T) {
	type testScenario struct {
		name           string
		baseStr        string
		otherStr       string
		expectedRelStr string
		expectErr      bool
		errContains    string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		base, err := Parse(s.baseStr)
		require.NoError(t, err)
		other, err := Parse(s.otherStr)
		require.NoError(t, err)

		rel, err := base.ExtractRelative(other)

		if s.expectErr {
			require.Error(t, err)
			require.ErrorIs(t, err, ErrNavigationError)
			require.Contains(t, err.Error(), s.errContains)
			return
		}

		require.NoError(t, err)
		expectedRel, err := Parse(s.expectedRelStr)
		require.NoError(t, err)
		require.Equal(t, expectedRel, rel, "Expected relative path '%s', but got '%s'", expectedRel.String(), rel.String())
	}

	t.Run("simple extraction", func(t *testing.T) {
		check(t, testScenario{
			baseStr:        "a/b",
			otherStr:       "a/b/c/d",
			expectedRelStr: "c/d",
		})
	})

	t.Run("simple extraction with trailing slash on other", func(t *testing.T) {
		check(t, testScenario{
			baseStr:        "a/b",
			otherStr:       "a/b/c/d/",
			expectedRelStr: "c/d/",
		})
	})

	t.Run("simple extraction with trailing slash on other (base has trailing slash)", func(t *testing.T) {
		check(t, testScenario{
			baseStr:        "a/b/",
			otherStr:       "a/b/c/d/",
			expectedRelStr: "c/d/",
		})
	})

	t.Run("identical paths", func(t *testing.T) {
		check(t, testScenario{
			baseStr:        "a/b",
			otherStr:       "a/b",
			expectedRelStr: ".", // Represents an empty path
		})
	})

	t.Run("identical paths with trailing slashes", func(t *testing.T) {
		check(t, testScenario{
			baseStr:        "a/b/",
			otherStr:       "a/b/",
			expectedRelStr: ".",
		})
	})

	t.Run("other path is not relative to base path when base has trailing slash", func(t *testing.T) {
		check(t, testScenario{
			baseStr:     "a/b/",
			otherStr:    "a/b",
			expectErr:   true,
			errContains: "navigation error: 'a/b' is not relative to 'a/b/'",
		})
	})

	t.Run("absolute paths", func(t *testing.T) {
		check(t, testScenario{
			baseStr:        "/a/b",
			otherStr:       "/a/b/c",
			expectedRelStr: "c",
		})
	})

	t.Run("root path as base", func(t *testing.T) {
		check(t, testScenario{
			baseStr:        "/",
			otherStr:       "/a/b",
			expectedRelStr: "a/b",
		})
	})

	t.Run("relative paths with parents", func(t *testing.T) {
		check(t, testScenario{
			baseStr:        "../a",
			otherStr:       "../a/b/c",
			expectedRelStr: "b/c",
		})
	})

	t.Run("error when other path is not a descendant of base", func(t *testing.T) {
		check(t, testScenario{
			baseStr:     "a/c",
			otherStr:    "a/b/d",
			expectErr:   true,
			errContains: "'a/b/d' is not relative to 'a/c'",
		})
	})

	t.Run("error when base path is longer than other path", func(t *testing.T) {
		check(t, testScenario{
			baseStr:     "a/b/c",
			otherStr:    "a/b",
			expectErr:   true,
			errContains: "'a/b' is not relative to 'a/b/c'",
		})
	})

	t.Run("error for absolute vs relative paths", func(t *testing.T) {
		check(t, testScenario{
			baseStr:     "/a/b",
			otherStr:    "a/b/c",
			expectErr:   true,
			errContains: "'a/b/c' is not relative to '/a/b'",
		})
	})
}
