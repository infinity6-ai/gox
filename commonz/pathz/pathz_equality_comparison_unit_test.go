package pathz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/stretchr/testify/require"
)

func TestUnitPathEquality(t *testing.T) {
	type testScenario struct {
		name       string
		path1      *pathz.Path
		path2      *pathz.Path
		expectEqual bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		require.Equal(t, s.expectEqual, s.path1.Equals(s.path2), "Equality mismatch for %s", s.name)
		// Check for symmetry
		require.Equal(t, s.expectEqual, s.path2.Equals(s.path1), "Symmetry mismatch for %s", s.name)
	}

	t.Run("identical absolute paths", func(t *testing.T) {
		p1 := pathz.New(-1, []string{"a", "b", "c"}, false)
		p2 := pathz.New(-1, []string{"a", "b", "c"}, false)
		check(t, testScenario{
			name:       "identical absolute paths",
			path1:      p1,
			path2:      p2,
			expectEqual: true,
		})
	})

	t.Run("different absolute paths", func(t *testing.T) {
		p1 := pathz.New(-1, []string{"a", "b", "c"}, false)
		p2 := pathz.New(-1, []string{"a", "b", "d"}, false)
		check(t, testScenario{
			name:       "different absolute paths",
			path1:      p1,
			path2:      p2,
			expectEqual: false,
		})
	})

	t.Run("identical relative paths", func(t *testing.T) {
		p1 := pathz.New(0, []string{"a", "b"}, false)
		p2 := pathz.New(0, []string{"a", "b"}, false)
		check(t, testScenario{
			name:       "identical relative paths",
			path1:      p1,
			path2:      p2,
			expectEqual: true,
		})
	})

	t.Run("different relative paths", func(t *testing.T) {
		p1 := pathz.New(0, []string{"a", "b"}, false)
		p2 := pathz.New(0, []string{"x", "y"}, false)
		check(t, testScenario{
			name:       "different relative paths",
			path1:      p1,
			path2:      p2,
			expectEqual: false,
		})
	})

	t.Run("same parts, different parents count", func(t *testing.T) {
		p1 := pathz.New(0, []string{"a", "b"}, false) // relative
		p2 := pathz.New(-1, []string{"a", "b"}, false) // absolute
		check(t, testScenario{
			name:       "same parts, different parents count",
			path1:      p1,
			path2:      p2,
			expectEqual: false,
		})
	})

	t.Run("same parents count, different parts length", func(t *testing.T) {
		p1 := pathz.New(0, []string{"a", "b"}, false)
		p2 := pathz.New(0, []string{"a"}, false)
		check(t, testScenario{
			name:       "same parents count, different parts length",
			path1:      p1,
			path2:      p2,
			expectEqual: false,
		})
	})

	t.Run("one path has ending slash, other does not", func(t *testing.T) {
		p1 := pathz.New(0, []string{"a", "b"}, false)
		p2 := pathz.New(0, []string{"a", "b"}, true)
		check(t, testScenario{
			name:       "one path has ending slash, other does not",
			path1:      p1,
			path2:      p2,
			expectEqual: false,
		})
	})

	t.Run("both paths have ending slash", func(t *testing.T) {
		p1 := pathz.New(0, []string{"a", "b"}, true)
		p2 := pathz.New(0, []string{"a", "b"}, true)
		check(t, testScenario{
			name:       "both paths have ending slash",
			path1:      p1,
			path2:      p2,
			expectEqual: true,
		})
	})

	t.Run("empty paths", func(t *testing.T) {
		p1 := pathz.New(0, nil, false)
		p2 := pathz.New(0, nil, false)
		check(t, testScenario{
			name:       "empty paths",
			path1:      p1,
			path2:      p2,
			expectEqual: true,
		})
	})

	t.Run("nil paths", func(t *testing.T) {
		check(t, testScenario{
			name:       "both nil paths",
			path1:      nil,
			path2:      nil,
			expectEqual: true,
		})
		check(t, testScenario{
			name:       "one nil path",
			path1:      pathz.New(0, nil, false),
			path2:      nil,
			expectEqual: false,
		})
		check(t, testScenario{
			name:       "other nil path",
			path1:      nil,
			path2:      pathz.New(0, nil, false),
			expectEqual: false,
		})
	})

	t.Run("path with different escape parents", func(t *testing.T) {
		p1 := pathz.New(1, []string{"a", "b"}, false) // ../a/b
		p2 := pathz.New(2, []string{"a", "b"}, false) // ../../a/b
		check(t, testScenario{
			name:       "path with different escape parents",
			path1:      p1,
			path2:      p2,
			expectEqual: false,
		})
	})
}

func TestUnitPathComparison(t *testing.T) {
	type testScenario struct {
		name     string
		path1    *pathz.Path
		path2    *pathz.Path
		expected int // -1 if path1 < path2, 0 if path1 == path2, 1 if path1 > path2
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got := s.path1.Compare(s.path2)
		require.Equal(t, s.expected, got, "Comparison mismatch for %s: %s vs %s", s.name, s.path1.String(), s.path2.String())
	}

	p := func(s string) *pathz.Path {
		ret, err := pathz.Parse(s)
		require.NoError(t, err)
		return ret
	}

	t.Run("identical paths", func(t *testing.T) {
		check(t, testScenario{
			name:     "identical paths",
			path1:    p("/a/b/c"),
			path2:    p("/a/b/c"),
			expected: 0,
		})
	})

	t.Run("path1 smaller than path2", func(t *testing.T) {
		check(t, testScenario{
			name:     "path1 smaller than path2",
			path1:    p("/a/b"),
			path2:    p("/a/b/c"),
			expected: -1,
		})
	})

	t.Run("path1 larger than path2", func(t *testing.T) {
		check(t, testScenario{
			name:     "path1 larger than path2",
			path1:    p("/a/b/c"),
			path2:    p("/a/b"),
			expected: 1,
		})
	})

	t.Run("different schemes (absolute vs relative)", func(t *testing.T) {
		check(t, testScenario{
			name:     "different schemes (absolute vs relative)",
			path1:    p("/a/b/c"), // "/a/b/c"
			path2:    p("a/b/c"),  // "a/b/c"
			expected: -1, // "/" < "a"
		})
	})

	t.Run("paths with trailing slashes", func(t *testing.T) {
		check(t, testScenario{
			name:     "path with trailing slash is smaller if same base path",
			path1:    p("/a/b"),
			path2:    p("/a/b/"),
			expected: -1, // "/a/b" < "/a/b/"
		})
		check(t, testScenario{
			name:     "path with trailing slash is larger if same base path",
			path1:    p("/a/b/"),
			path2:    p("/a/b"),
			expected: 1, // "/a/b/" > "/a/b"
		})
	})

	t.Run("paths with parent navigation", func(t *testing.T) {
		check(t, testScenario{
			name:     "paths with parent navigation - equal",
			path1:    p("../a"),
			path2:    p("../a"),
			expected: 0,
		})
		check(t, testScenario{
			name:     "paths with parent navigation - different levels",
			path1:    p("../a"),   // "../a"
			path2:    p("../../a"), // "../../a"
			expected: 1, // "../a" > "../../a"
		})
		check(t, testScenario{
			name:     "paths with parent navigation - different levels, reversed",
			path1:    p("../../a"), // "../../a"
			path2:    p("../a"),   // "../a"
			expected: -1, // "../../a" < "../a"
		})
	})

	t.Run("empty path vs non-empty", func(t *testing.T) {
		check(t, testScenario{
			name:     "empty path vs non-empty",
			path1:    p(""),
			path2:    p("a"),
			expected: -1, // "" < "a"
		})
		check(t, testScenario{
			name:     "non-empty vs empty path",
			path1:    p("a"),
			path2:    p(""),
			expected: 1, // "a" > ""
		})
	})

	t.Run("empty path vs root path", func(t *testing.T) {
		check(t, testScenario{
			name:     "empty path vs root path",
			path1:    p(""), // ""
			path2:    p("/"), // "/"
			expected: -1, // "" < "/"
		})
		check(t, testScenario{
			name:     "root path vs empty path",
			path1:    p("/"), // "/"
			path2:    p(""),  // ""
			expected: 1, // "/" > ""
		})
	})

	t.Run("complex comparison", func(t *testing.T) {
		check(t, testScenario{
			name:     "complex comparison 1",
			path1:    p("/a/b/z"),
			path2:    p("/a/b/c"),
			expected: 1,
		})
		check(t, testScenario{
			name:     "complex comparison 2",
			path1:    p("b/c"),
			path2:    p("a/b/c"),
			expected: 1,
		})
		check(t, testScenario{
			name:     "complex comparison 3",
			path1:    p("a/b/c"),
			path2:    p("b/c"),
			expected: -1,
		})
	})
}
