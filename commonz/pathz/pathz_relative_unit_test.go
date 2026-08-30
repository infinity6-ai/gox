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

	scenarios := []testScenario{
		{
			name:           "simple extraction",
			baseStr:        "a/b",
			otherStr:       "a/b/c/d",
			expectedRelStr: "c/d",
		},
		{
			name:           "simple extraction with trailing slash on other",
			baseStr:        "a/b",
			otherStr:       "a/b/c/d/",
			expectedRelStr: "c/d/",
		},
		{
			name:           "simple extraction with trailing slash on other",
			baseStr:        "a/b/",
			otherStr:       "a/b/c/d/",
			expectedRelStr: "c/d/",
		},
		{
			name:           "identical paths",
			baseStr:        "a/b",
			otherStr:       "a/b",
			expectedRelStr: ".", // Represents an empty path
		},
		{
			name:           "identical paths with trailing slashes",
			baseStr:        "a/b/",
			otherStr:       "a/b/",
			expectedRelStr: ".",
		},
		{
			name:           "identical paths with trailing slashes",
			baseStr:        "a/b/",
			otherStr:       "a/b",
			expectedRelStr: ".",
			expectErr:      true,
			errContains:    "navigation error: 'a/b' is not relative to 'a/b/'",
		},
		{
			name:           "absolute paths",
			baseStr:        "/a/b",
			otherStr:       "/a/b/c",
			expectedRelStr: "c",
		},
		{
			name:           "root path as base",
			baseStr:        "/",
			otherStr:       "/a/b",
			expectedRelStr: "a/b",
		},
		{
			name:           "relative paths with parents",
			baseStr:        "../a",
			otherStr:       "../a/b/c",
			expectedRelStr: "b/c",
		},
		{
			name:        "error when not a base",
			baseStr:     "a/c",
			otherStr:    "a/b/d",
			expectErr:   true,
			errContains: "'a/b/d' is not relative to 'a/c'",
		},
		{
			name:        "error when base is longer",
			baseStr:     "a/b/c",
			otherStr:    "a/b",
			expectErr:   true,
			errContains: "'a/b' is not relative to 'a/b/c'",
		},
		{
			name:        "error for absolute vs relative",
			baseStr:     "/a/b",
			otherStr:    "a/b/c",
			expectErr:   true,
			errContains: "'a/b/c' is not relative to '/a/b'",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			check(t, s)
		})
	}
}
