package pathz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitPathGetters(t *testing.T) {
	type testScenario struct {
		pathStr      string
		isAbsolute   bool
		isContained  bool
		isEscaped    bool
		partSlice    []string
		base         string
		dirPathStr   string
		parentIsNull bool // New field to indicate if the parent should be nil
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

		// Test Dir() first to get expected parent path
		expectedDirPath, err := Parse(s.dirPathStr)
		require.NoError(t, err)
		dir := p.Dir()
		if s.parentIsNull {
			require.Nil(t, dir, "Dir() should be nil for %q", s.pathStr)
		} else {
			require.Equal(t, expectedDirPath, dir, "Dir() path mismatch for %q", s.pathStr)
		}

		// Test Parent()
		parent, base := p.Parent()
		require.Equal(t, s.base, base, "Parent() base mismatch for %q", s.pathStr)
		if s.parentIsNull {
			require.Nil(t, parent, "Parent() should be nil for %q", s.pathStr)
		} else {
			require.Equal(t, expectedDirPath, parent, "Parent() path mismatch for %q", s.pathStr)
		}
	}

	t.Run("absolute path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:     "/a/b/c",
			isAbsolute:  true,
			isContained: false,
			isEscaped:   false,
			partSlice:   []string{"a", "b", "c"},
			base:        "c",
			dirPathStr:  "/a/b",
		})
	})

	t.Run("contained path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:     "a/b/c",
			isAbsolute:  false,
			isContained: true,
			isEscaped:   false,
			partSlice:   []string{"a", "b", "c"},
			base:        "c",
			dirPathStr:  "a/b",
		})
	})

	t.Run("escaped path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:     "../a/b",
			isAbsolute:  false,
			isContained: false,
			isEscaped:   true,
			partSlice:   []string{"a", "b"},
			base:        "b",
			dirPathStr:  "../a",
		})
	})

	t.Run("doubled escaped path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:     "../../a/b",
			isAbsolute:  false,
			isContained: false,
			isEscaped:   true,
			partSlice:   []string{"a", "b"},
			base:        "b",
			dirPathStr:  "../../a",
		})
	})

	t.Run("empty path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:     "",
			isAbsolute:  false,
			isContained: true,
			isEscaped:   false,
			partSlice:   []string{},
			base:        "",
			dirPathStr:  "",
		})
	})

	t.Run("root path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:     "/",
			isAbsolute:  true,
			isContained: false,
			isEscaped:   false,
			partSlice:   []string{},
			base:        "",
			dirPathStr:  "/",
		})
	})

	t.Run("single element path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:      "a",
			isAbsolute:   false,
			isContained:  true,
			isEscaped:    false,
			partSlice:    []string{"a"},
			base:         "a",
			dirPathStr:   "",
			parentIsNull: true,
		})
	})

	t.Run("dot path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:     ".",
			isAbsolute:  false,
			isContained: true,
			isEscaped:   false,
			partSlice:   []string{},
			base:        "",
			dirPathStr:  "",
		})
	})

	t.Run("dot slash path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:     "./",
			isAbsolute:  false,
			isContained: true,
			isEscaped:   false,
			partSlice:   []string{},
			base:        "",
			dirPathStr:  "",
		})
	})

	t.Run("path ending with slash", func(t *testing.T) {
		check(t, testScenario{
			pathStr:     "a/b/",
			isAbsolute:  false,
			isContained: true,
			isEscaped:   false,
			partSlice:   []string{"a", "b"},
			base:        "b",
			dirPathStr:  "a",
		})
	})

	t.Run("single element path 'bla'", func(t *testing.T) {
		check(t, testScenario{
			pathStr:      "bla",
			isAbsolute:   false,
			isContained:  true,
			isEscaped:    false,
			partSlice:    []string{"bla"},
			base:         "bla",
			dirPathStr:   "",
			parentIsNull: true,
		})
	})

	t.Run("complex path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:     "/a/b/../c/",
			isAbsolute:  true,
			isContained: false,
			isEscaped:   false,
			partSlice:   []string{"a", "c"},
			base:        "c",
			dirPathStr:  "/a",
		})
	})
}

func TestUnitParentMethod(t *testing.T) {
	type testScenario struct {
		pathStr       string
		parentPathStr string
		parentBase    string
		parentIsNull  bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		p, err := Parse(s.pathStr)
		require.NoError(t, err)

		parent, base := p.Parent()
		require.Equal(t, s.parentBase, base, "Parent() base mismatch for %q", s.pathStr)

		if s.parentIsNull {
			require.Nil(t, parent, "Parent() should be nil for %q", s.pathStr)
		} else {
			expectedParentPath, err := Parse(s.parentPathStr)
			require.NoError(t, err)
			require.Equal(t, expectedParentPath, parent, "Parent() path mismatch for %q", s.pathStr)
		}
	}

	t.Run("path with multiple parts", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "a/b/c",
			parentPathStr: "a/b",
			parentBase:    "c",
		})
	})

	t.Run("path with two parts", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "a/b",
			parentPathStr: "a",
			parentBase:    "b",
		})
	})

	t.Run("single element path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:      "a",
			parentBase:   "a",
			parentIsNull: true,
		})
	})

	t.Run("empty path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "",
			parentPathStr: "",
			parentBase:    "",
		})
	})

	t.Run("root path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "/",
			parentPathStr: "/",
			parentBase:    "",
		})
	})

	t.Run("absolute path with one part", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "/a",
			parentPathStr: "/",
			parentBase:    "a",
		})
	})

	t.Run("escaped path", func(t *testing.T) {
		check(t, testScenario{
			pathStr:       "../a/b",
			parentPathStr: "../a",
			parentBase:    "b",
		})
	})
}
