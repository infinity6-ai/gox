package pathz_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/pathz"
)

func TestUnitParse(t *testing.T) {
	type testScenario struct {
		input                  string
		expectedParts          []string
		expectedParents        int
		expectedHasEndingSlash bool
		expectedError          string
		expectedPanic          string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		var p *pathz.Path
		var err error
		if s.expectedPanic != "" {
			require.PanicsWithValue(t, s.expectedPanic, func() {
				p, err = pathz.Parse(s.input)
			})
			require.Nil(t, p)
			require.Nil(t, err)
			return
		}
		pErr := errorz.Unpanic(func() {
			p, err = pathz.Parse(s.input)
		})
		require.Nil(t, pErr)

		if s.expectedError != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedError)
			require.Nil(t, p)
		} else {
			require.NoError(t, err)
			require.NotNil(t, p)
			require.Equal(t, s.expectedParts, p.Parts())
			require.Equal(t, s.expectedParents, p.Parents())
			require.Equal(t, s.expectedHasEndingSlash, p.HasEndingSlash())
		}
	}

	t.Run("absolute path", func(t *testing.T) {
		check(t, testScenario{
			input:           "/a/b/c",
			expectedParts:   []string{"a", "b", "c"},
			expectedParents: -1,
		})
	})

	t.Run("relative path", func(t *testing.T) {
		check(t, testScenario{
			input:           "a/b/c",
			expectedParts:   []string{"a", "b", "c"},
			expectedParents: 0,
		})
	})

	t.Run("path with current dir", func(t *testing.T) {
		check(t, testScenario{
			input:           "a/./b/c",
			expectedParts:   []string{"a", "b", "c"},
			expectedParents: 0,
		})
	})

	t.Run("path with parent dir", func(t *testing.T) {
		check(t, testScenario{
			input:           "a/b/../c",
			expectedParts:   []string{"a", "c"},
			expectedParents: 0,
		})
	})

	t.Run("path starting with parent dir (relative)", func(t *testing.T) {
		check(t, testScenario{
			input:           "../a/b",
			expectedParts:   []string{"a", "b"},
			expectedParents: 1,
		})
	})

	t.Run("path starting with parent dir (absolute)", func(t *testing.T) {
		check(t, testScenario{
			input:           "/../a/b",
			expectedParts:   []string{"a", "b"},
			expectedParents: -1,
		})
	})

	t.Run("path with multiple parent dir navigations", func(t *testing.T) {
		check(t, testScenario{
			input:           "a/b/c/../../d",
			expectedParts:   []string{"a", "d"},
			expectedParents: 0,
		})
	})

	t.Run("path with excessive parent dir navigations", func(t *testing.T) {
		check(t, testScenario{
			input:           "a/../../b",
			expectedParts:   []string{"b"},
			expectedParents: 1,
		})
	})

	t.Run("empty path", func(t *testing.T) {
		check(t, testScenario{
			input:           "",
			expectedParts:   nil,
			expectedParents: 0,
		})
	})

	t.Run("path with only slashes", func(t *testing.T) {
		check(t, testScenario{
			input:                  "///",
			expectedParts:          nil,
			expectedParents:        -1,
			expectedHasEndingSlash: true,
		})
	})

	t.Run("path with only current dir", func(t *testing.T) {
		check(t, testScenario{
			input:           "././.",
			expectedParts:   nil,
			expectedParents: 0,
		})
	})

	t.Run("path with only parent dir", func(t *testing.T) {
		check(t, testScenario{
			input:                  "../../",
			expectedParts:          []string{},
			expectedParents:        2,
			expectedHasEndingSlash: true,
		})
	})

	t.Run("path with illegal characters - null byte", func(t *testing.T) {
		check(t, testScenario{
			input:         "a\x00b/c",
			expectedError: "path contains illegal character 1 '\x00' in 'a\x00b/c'",
		})
	})

	t.Run("path with illegal characters - control character", func(t *testing.T) {
		check(t, testScenario{
			input:         "a\tb/c",
			expectedError: "path contains illegal character 1 '\t' in 'a\tb/c'",
		})
	})

	t.Run("path with trailing slash", func(t *testing.T) {
		check(t, testScenario{
			input:                  "a/b/c/",
			expectedParts:          []string{"a", "b", "c"},
			expectedParents:        0,
			expectedHasEndingSlash: true,
		})
	})

	t.Run("path with leading slash", func(t *testing.T) {
		check(t, testScenario{
			input:           "/a/b/c",
			expectedParts:   []string{"a", "b", "c"},
			expectedParents: -1,
		})
	})

	t.Run("path with mixed slashes", func(t *testing.T) {
		check(t, testScenario{
			input:           "//a/b///c",
			expectedParts:   []string{"a", "b", "c"},
			expectedParents: -1,
		})
	})

	t.Run("complex path", func(t *testing.T) {
		check(t, testScenario{
			input:           "/./a/../b/./c/../../d/e/.",
			expectedParts:   []string{"d", "e"},
			expectedParents: -1,
		})
	})

	t.Run("path with illegal component dots", func(t *testing.T) {
		check(t, testScenario{
			input:         "a/.../b",
			expectedError: "path contains illegal component 1: \"...\"",
		})
	})

	t.Run("path with illegal component more dots", func(t *testing.T) {
		check(t, testScenario{
			input:         "a/..../b",
			expectedError: "path contains illegal component 1: \"....\"",
		})
	})

	t.Run("path with leading illegal component dots", func(t *testing.T) {
		check(t, testScenario{
			input:         ".../b",
			expectedError: "path contains illegal component 0: \"...\"",
		})
	})

	t.Run("path with trailing illegal component dots", func(t *testing.T) {
		check(t, testScenario{
			input:         "a/...",
			expectedError: "path contains illegal component 1: \"...\"",
		})
	})

	t.Run("path /", func(t *testing.T) {
		check(t, testScenario{
			input:                  "/",
			expectedHasEndingSlash: false,
			expectedParents:        -1,
		})
	})

	t.Run("path /x/*", func(t *testing.T) {
		check(t, testScenario{
			input:                  "/x/*",
			expectedParts:          []string{"x", "*"},
			expectedHasEndingSlash: false,
			expectedParents:        -1,
		})
	})

	t.Run("path /*", func(t *testing.T) {
		check(t, testScenario{
			input:                  "/*",
			expectedParts:          []string{"*"},
			expectedHasEndingSlash: false,
			expectedParents:        -1,
		})
	})
}
