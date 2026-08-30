package pathz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitPathGetters(t *testing.T) {
	type testScenario struct {
		pathStr       string
		isAbsolute    bool
		isContained   bool
		isEscaped     bool
		partSlice     []string
		base          string
		parentPathStr string
		parentBase    string
		dirPathStr    string
		parentIsNull  bool // New field to indicate if the parent should be nil
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		p, err := Parse(s.pathStr)
		require.NoError(t, err)

		require.Equal(t, s.isAbsolute, p.IsAbsolute(), "IsAbsolute mismatch for %q", s.pathStr)
		require.Equal(t, s.isContained, p.IsContained(), "IsContained mismatch for %q", s.pathStr)
		require.Equal(t, s.isEscaped, p.IsEscaped(), "IsEscaped mismatch for %q", s.pathStr)
		require.Equal(t, s.partSlice, p.Split(), "PartSlice mismatch for %q", s.pathStr)
		require.Equal(t, s.base, p.Base(), "Base mismatch for %q", s.pathStr)

		// Test Parent()
		parent, base := p.Parent()
		require.Equal(t, s.parentBase, base, "Parent() base mismatch for %q", s.pathStr)
		if s.parentIsNull {
			require.Nil(t, parent, "Parent() should be nil for %q", s.pathStr)
		} else {
			expectedParentPath, err := Parse(s.parentPathStr)
			require.NoError(t, err)
			require.Equal(t, expectedParentPath, parent, "Parent() path mismatch for %q", s.pathStr)
		}

		// Test Dir()
		dir := p.Dir()
		expectedDirPath, err := Parse(s.dirPathStr)
		require.NoError(t, err)
		require.Equal(t, expectedDirPath, dir, "Dir() path mismatch for %q", s.pathStr)
	}

	t.Run("absolute path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "/a/b/c",
			isAbsolute:    true,
			isContained:   false,
			isEscaped:     false,
			partSlice:     []string{"a", "b", "c"},
			base:          "c",
			parentPathStr: "/a/b",
			parentBase:    "c",
			dirPathStr:    "/a/b",
		})
	})

	t.Run("contained path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "a/b/c",
			isAbsolute:    false,
			isContained:   true,
			isEscaped:     false,
			partSlice:     []string{"a", "b", "c"},
			base:          "c",
			parentPathStr: "a/b",
			parentBase:    "c",
			dirPathStr:    "a/b",
		})
	})

	t.Run("escaped path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "../a/b",
			isAbsolute:    false,
			isContained:   false,
			isEscaped:     true,
			partSlice:     []string{"a", "b"},
			base:          "b",
			parentPathStr: "../a",
			parentBase:    "b",
			dirPathStr:    "../a",
		})
	})

	t.Run("doubled escaped path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "../../a/b",
			isAbsolute:    false,
			isContained:   false,
			isEscaped:     true,
			partSlice:     []string{"a", "b"},
			base:          "b",
			parentPathStr: "../../a",
			parentBase:    "b",
			dirPathStr:    "../../a",
		})
	})

	t.Run("empty path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "",
			isAbsolute:    false,
			isContained:   true,
			isEscaped:     false,
			partSlice:     []string{},
			base:          "",
			parentPathStr: "",
			parentBase:    "",
			dirPathStr:    "",
		})
	})

	t.Run("root path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "/",
			isAbsolute:    true,
			isContained:   false,
			isEscaped:     false,
			partSlice:     []string{},
			base:          "",
			parentPathStr: "/",
			parentBase:    "",
			dirPathStr:    "/",
		})
	})

	t.Run("single element path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "a",
			isAbsolute:    false,
			isContained:   true,
			isEscaped:     false,
			partSlice:     []string{"a"},
			base:          "a",
			parentPathStr: "",
			parentBase:    "a",
			dirPathStr:    "",
			parentIsNull:  true,
		})
	})

	t.Run("dot path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       ".",
			isAbsolute:    false,
			isContained:   true,
			isEscaped:     false,
			partSlice:     []string{},
			base:          "",
			parentPathStr: "",
			parentBase:    "",
			dirPathStr:    "",
		})
	})

	t.Run("dot slash path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "./",
			isAbsolute:    false,
			isContained:   true,
			isEscaped:     false,
			partSlice:     []string{},
			base:          "",
			parentPathStr: "",
			parentBase:    "",
			dirPathStr:    "",
		})
	})

	t.Run("path ending with slash", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "a/b/",
			isAbsolute:    false,
			isContained:   true,
			isEscaped:     false,
			partSlice:     []string{"a", "b"},
			base:          "b",
			parentPathStr: "a",
			parentBase:    "b",
			dirPathStr:    "a",
		})
	})

	t.Run("single element path 'bla'", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "bla",
			isAbsolute:    false,
			isContained:   true,
			isEscaped:     false,
			partSlice:     []string{"bla"},
			base:          "bla",
			parentPathStr: "",
			parentBase:    "bla",
			dirPathStr:    "",
			parentIsNull:  true,
		})
	})

	t.Run("complex path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "/a/b/../c/",
			isAbsolute:    true,
			isContained:   false,
			isEscaped:     false,
			partSlice:     []string{"a", "c"},
			base:          "c",
			parentPathStr: "/a",
			parentBase:    "c",
			dirPathStr:    "/a",
		})
	})
}
