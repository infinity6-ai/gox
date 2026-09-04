package pathz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitPathGetters(t *testing.T) {
	type testScenario struct {
		path           *Path
		pathStr        string // For error messages
		isAbsolute     bool
		isContained    bool
		isEscaped      bool
		partSlice      []string
		base           string
		parent         *Path // Expected parent path, nil if no parent
		hasEndingSlash bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		p := s.path

		require.Equal(t, s.isAbsolute, p.IsAbsolute(), "IsAbsolute mismatch for %q", s.pathStr)
		require.Equal(t, s.isContained, p.IsContained(), "IsContained mismatch for %q", s.pathStr)
		require.Equal(t, s.isEscaped, p.IsEscaped(), "IsEscaped mismatch for %q", s.pathStr)
		require.Equal(t, s.partSlice, p.Split(), "PartSlice mismatch for %q", s.pathStr)
		require.Equal(t, s.base, p.Base(), "Base mismatch for %q", s.pathStr)

		dir := p.Dir()
		if s.parent == nil {
			require.Nil(t, dir, "Dir() should be nil for %q", s.pathStr)
		} else {
			require.Equal(t, s.parent, dir, "Dir() path mismatch for %q", s.pathStr)
		}

		// Test Parent()
		parent, base, hasEndingSlash := p.Parent()
		require.Equal(t, s.hasEndingSlash, hasEndingSlash, "Parent() hasEndingSlash mismatch for %q", s.pathStr)
		require.Equal(t, s.base, base, "Parent() base mismatch for %q", s.pathStr)
		if s.parent == nil {
			require.Nil(t, parent, "Parent() should be nil for %q", s.pathStr)
		} else {
			require.Equal(t, s.parent, parent, "Parent() path mismatch for %q", s.pathStr)
		}

		// Test BaseSlash()
		baseSlash, hasEndingSlash := p.BaseSlash()
		require.Equal(t, s.hasEndingSlash, hasEndingSlash, "BaseSlash() hasEndingSlash mismatch for %q", s.pathStr)
		require.Equal(t, s.base, baseSlash, "BaseSlash() base mismatch for %q", s.pathStr)
	}

	t.Run("absolute path", func(t *testing.T) {
		check(t, testScenario{
			path:           New(-1, []string{"a", "b", "c"}, false),
			pathStr:        "/a/b/c",
			isAbsolute:     true,
			isContained:    false,
			isEscaped:      false,
			partSlice:      []string{"a", "b", "c"},
			base:           "c",
			parent:         New(-1, []string{"a", "b"}, false),
			hasEndingSlash: false,
		})
	})

	t.Run("contained path", func(t *testing.T) {
		check(t, testScenario{
			path:           New(0, []string{"a", "b", "c"}, false),
			pathStr:        "a/b/c",
			isAbsolute:     false,
			isContained:    true,
			isEscaped:      false,
			partSlice:      []string{"a", "b", "c"},
			base:           "c",
			parent:         New(0, []string{"a", "b"}, false),
			hasEndingSlash: false,
		})
	})

	t.Run("escaped path", func(t *testing.T) {
		check(t, testScenario{
			path:           New(1, []string{"a", "b"}, false),
			pathStr:        "../a/b",
			isAbsolute:     false,
			isContained:    false,
			isEscaped:      true,
			partSlice:      []string{"a", "b"},
			base:           "b",
			parent:         New(1, []string{"a"}, false),
			hasEndingSlash: false,
		})
	})

	t.Run("doubled escaped path", func(t *testing.T) {
		check(t, testScenario{
			path:           New(2, []string{"a", "b"}, false),
			pathStr:        "../../a/b",
			isAbsolute:     false,
			isContained:    false,
			isEscaped:      true,
			partSlice:      []string{"a", "b"},
			base:           "b",
			parent:         New(2, []string{"a"}, false),
			hasEndingSlash: false,
		})
	})

	t.Run("escaped path with single part", func(t *testing.T) {
		check(t, testScenario{
			path:           New(2, []string{"a"}, false),
			pathStr:        "../../a",
			isAbsolute:     false,
			isContained:    false,
			isEscaped:      true,
			partSlice:      []string{"a"},
			base:           "a",
			parent:         New(2, []string{}, false), // Path for "../.."
			hasEndingSlash: false,
		})
	})

	t.Run("doubly escaped path only", func(t *testing.T) {
		check(t, testScenario{
			path:           New(2, []string{}, false),
			pathStr:        "../..",
			isAbsolute:     false,
			isContained:    false,
			isEscaped:      true,
			partSlice:      []string{},
			base:           "",
			parent:         New(1, []string{}, false), // Path for ".."
			hasEndingSlash: false,
		})
	})

	t.Run("single escaped path only", func(t *testing.T) {
		check(t, testScenario{
			path:           New(1, []string{}, false),
			pathStr:        "..",
			isAbsolute:     false,
			isContained:    false,
			isEscaped:      true,
			partSlice:      []string{},
			base:           "",
			parent:         nil,
			hasEndingSlash: false,
		})
	})

	t.Run("empty path", func(t *testing.T) {
		check(t, testScenario{
			path:           New(0, []string{}, false),
			pathStr:        "",
			isAbsolute:     false,
			isContained:    true,
			isEscaped:      false,
			partSlice:      []string{},
			base:           "",
			parent:         nil, // Dir() is nil, Parent() is (nil, "")
			hasEndingSlash: false,
		})
	})

	t.Run("root path", func(t *testing.T) {
		check(t, testScenario{
			path:           New(-1, []string{}, false),
			pathStr:        "/",
			isAbsolute:     true,
			isContained:    false,
			isEscaped:      false,
			partSlice:      []string{},
			base:           "",
			parent:         nil,
			hasEndingSlash: false,
		})
	})

	t.Run("single element path", func(t *testing.T) {
		check(t, testScenario{
			path:           New(0, []string{"a"}, false),
			pathStr:        "a",
			isAbsolute:     false,
			isContained:    true,
			isEscaped:      false,
			partSlice:      []string{"a"},
			base:           "a",
			parent:         nil,
			hasEndingSlash: false,
		})
	})

	t.Run("dot path", func(t *testing.T) {
		check(t, testScenario{
			path:           New(0, []string{}, false),
			pathStr:        ".",
			isAbsolute:     false,
			isContained:    true,
			isEscaped:      false,
			partSlice:      []string{},
			base:           "",
			parent:         nil,
			hasEndingSlash: false,
		})
	})

	t.Run("dot slash path", func(t *testing.T) {
		check(t, testScenario{
			path:           New(0, []string{}, true),
			pathStr:        "./",
			isAbsolute:     false,
			isContained:    true,
			isEscaped:      false,
			partSlice:      []string{},
			base:           "",
			parent:         nil,
			hasEndingSlash: false,
		})
	})

	t.Run("path ending with slash", func(t *testing.T) {
		check(t, testScenario{
			path:           New(0, []string{"a", "b"}, true),
			pathStr:        "a/b/",
			isAbsolute:     false,
			isContained:    true,
			isEscaped:      false,
			partSlice:      []string{"a", "b"},
			base:           "b",
			parent:         New(0, []string{"a"}, false),
			hasEndingSlash: true,
		})
	})

	t.Run("single element path 'bla'", func(t *testing.T) {
		check(t, testScenario{
			path:           New(0, []string{"bla"}, false),
			pathStr:        "bla",
			isAbsolute:     false,
			isContained:    true,
			isEscaped:      false,
			partSlice:      []string{"bla"},
			base:           "bla",
			parent:         nil,
			hasEndingSlash: false,
		})
	})

	t.Run("cleaned absolute path with parent nav and ending slash", func(t *testing.T) {
		check(t, testScenario{
			path:           New(-1, []string{"a", "c"}, true),
			pathStr:        "/a/c/",
			isAbsolute:     true,
			isContained:    false,
			isEscaped:      false,
			partSlice:      []string{"a", "c"},
			base:           "c",
			parent:         New(-1, []string{"a"}, false),
			hasEndingSlash: true,
		})
	})
}

func TestUnitParentMethod(t *testing.T) {
	type testScenario struct {
		path           *Path
		pathStr        string
		expectedParent *Path
		parentBase     string
		hasEndingSlash bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		p := s.path

		parent, base, hasEndingSlash := p.Parent()
		require.Equal(t, s.hasEndingSlash, hasEndingSlash, "hasEndingSlash mismatch for %q", s.pathStr)
		require.Equal(t, s.parentBase, base, "Parent() base mismatch for %q", s.pathStr)

		if s.expectedParent == nil {
			require.Nil(t, parent, "Parent() should be nil for %q", s.pathStr)
		} else {
			require.Equal(t, s.expectedParent, parent, "Parent() path mismatch for %q", s.pathStr)
		}
	}

	t.Run("path with multiple parts", func(t *testing.T) {
		check(t, testScenario{
			path:           New(0, []string{"a", "b", "c"}, false),
			pathStr:        "a/b/c",
			expectedParent: New(0, []string{"a", "b"}, false),
			parentBase:     "c",
			hasEndingSlash: false,
		})
	})

	t.Run("path with two parts", func(t *testing.T) {
		check(t, testScenario{
			path:           New(0, []string{"a", "b"}, false),
			pathStr:        "a/b",
			expectedParent: New(0, []string{"a"}, false),
			parentBase:     "b",
			hasEndingSlash: false,
		})
	})

	t.Run("single element path", func(t *testing.T) {
		check(t, testScenario{
			path:           New(0, []string{"a"}, false),
			pathStr:        "a",
			expectedParent: nil,
			parentBase:     "a",
			hasEndingSlash: false,
		})
	})

	t.Run("empty path", func(t *testing.T) {
		check(t, testScenario{
			path:           New(0, []string{}, false),
			pathStr:        "",
			expectedParent: nil,
			parentBase:     "",
			hasEndingSlash: false,
		})
	})

	t.Run("root path", func(t *testing.T) {
		check(t, testScenario{
			path:           New(-1, []string{}, false),
			pathStr:        "/",
			expectedParent: nil,
			parentBase:     "",
			hasEndingSlash: false,
		})
	})

	t.Run("absolute path with one part", func(t *testing.T) {
		check(t, testScenario{
			path:           New(-1, []string{"a"}, false),
			pathStr:        "/a",
			expectedParent: nil,
			parentBase:     "a",
			hasEndingSlash: false,
		})
	})

	t.Run("escaped path with single part", func(t *testing.T) {
		check(t, testScenario{
			path:           New(2, []string{"a"}, false),
			pathStr:        "../../a",
			expectedParent: New(2, []string{}, false), // Path for "../.."
			parentBase:     "a",
			hasEndingSlash: false,
		})
	})

	t.Run("doubly escaped path only", func(t *testing.T) {
		check(t, testScenario{
			path:           New(2, []string{}, false),
			pathStr:        "../..",
			expectedParent: New(1, []string{}, false), // Path for ".."
			parentBase:     "",
			hasEndingSlash: false,
		})
	})

	t.Run("single escaped path only", func(t *testing.T) {
		check(t, testScenario{
			path:           New(1, []string{}, false),
			pathStr:        "..",
			expectedParent: nil,
			parentBase:     "",
			hasEndingSlash: false,
		})
	})

	t.Run("escaped path", func(t *testing.T) {
		check(t, testScenario{
			path:           New(1, []string{"a", "b"}, false),
			pathStr:        "../a/b",
			expectedParent: New(1, []string{"a"}, false),
			parentBase:     "b",
			hasEndingSlash: false,
		})
	})

	t.Run("path /x/*", func(t *testing.T) {
		check(t, testScenario{
			path:           New(-1, []string{"x", "*"}, false),
			pathStr:        "/x/*",
			expectedParent: New(-1, []string{"x"}, false),
			parentBase:     "*",
			hasEndingSlash: false,
		})
	})

	// t.Run("path /*", func(t *testing.T) {
	// 	check(t, testScenario{
	// 		path:           New(-1, []string{"*"}, false),
	// 		pathStr:        "/*",
	// 		expectedParent: New(-1, nil, false),
	// 		parentBase:     "*",
	// 		hasEndingSlash: false,
	// 	})
	// })
}
