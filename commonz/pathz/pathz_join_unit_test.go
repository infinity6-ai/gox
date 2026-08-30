package pathz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitPathJoin(t *testing.T) {
	type testScenario struct {
		name              string
		basePath          *Path
		others            []*Path
		expectedResult    *Path
		expectedContained *Path // Same reference as expectedResult if contained, otherwise nil
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		contained, result := s.basePath.Join(s.others...)

		require.Equal(t, s.expectedResult, result, "Result path mismatch for scenario: %s", s.name)

		if s.expectedContained == nil {
			require.Nil(t, contained, "Contained path should be nil for scenario: %s", s.name)
		} else {
			// This checks for reference equality as per the requirement
			require.True(t, contained == result, "Contained path should be the same reference as result for scenario: %s", s.name)
			require.Equal(t, s.expectedContained, contained, "Contained path value mismatch for scenario: %s", s.name)
		}
	}

	t.Run("join two relative paths", func(t *testing.T) {
		base := New(0, []string{"a", "b"}, false)
		other := New(0, []string{"c", "d"}, false)
		expected := New(0, []string{"a", "b", "c", "d"}, false)
		check(t, testScenario{
			name:              "a/b + c/d",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: expected,
		})
	})

	t.Run("join relative path with absolute path (absolute takes precedence)", func(t *testing.T) {
		base := New(0, []string{"a", "b"}, false)
		other := New(-1, []string{"c", "d"}, false)
		expected := New(-1, []string{"c", "d"}, false)
		check(t, testScenario{
			name:              "a/b + /c/d",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: nil, // Absolute path is not contained
		})
	})

	t.Run("join absolute path with relative path", func(t *testing.T) {
		base := New(-1, []string{"a", "b"}, false)
		other := New(0, []string{"c", "d"}, false)
		expected := New(-1, []string{"a", "b", "c", "d"}, false)
		check(t, testScenario{
			name:              "/a/b + c/d",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: nil,
		})
	})

	t.Run("join relative path with escaped path", func(t *testing.T) {
		base := New(0, []string{"a", "b"}, false)
		other := New(1, []string{"c", "d"}, false)         // ../c/d
		expected := New(0, []string{"a", "c", "d"}, false) // a/b/../c/d -> a/c/d
		check(t, testScenario{
			name:              "a/b + ../c/d",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: expected,
		})
	})

	t.Run("join escaped path with relative path", func(t *testing.T) {
		base := New(1, []string{"a", "b"}, false) // ../a/b
		other := New(0, []string{"c", "d"}, false)
		expected := New(1, []string{"a", "b", "c", "d"}, false) // ../a/b/c/d
		check(t, testScenario{
			name:              "../a/b + c/d",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: nil, // Escaped path is not contained
		})
	})

	t.Run("join with empty base path", func(t *testing.T) {
		base := New(0, nil, false) // ""
		other := New(0, []string{"a", "b"}, false)
		expected := New(0, []string{"a", "b"}, false)
		check(t, testScenario{
			name:              "'' + a/b",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: expected,
		})
	})

	t.Run("join with multiple others", func(t *testing.T) {
		base := New(0, []string{"a"}, false)
		other1 := New(0, []string{"b"}, false)
		other2 := New(0, []string{"c"}, false)
		expected := New(0, []string{"a", "b", "c"}, false)
		check(t, testScenario{
			name:              "a + b + c",
			basePath:          base,
			others:            []*Path{other1, other2},
			expectedResult:    expected,
			expectedContained: expected,
		})
	})

	t.Run("join with intermediate absolute path", func(t *testing.T) {
		base := New(0, []string{"a"}, false)
		other1 := New(0, []string{"b"}, false)
		other2 := New(-1, []string{"c"}, false)
		other3 := New(0, []string{"d"}, false)
		expected := New(-1, []string{"c", "d"}, false) // /c overrides a/b, then d joins
		check(t, testScenario{
			name:              "a + b + /c + d",
			basePath:          base,
			others:            []*Path{other1, other2, other3},
			expectedResult:    expected,
			expectedContained: nil,
		})
	})

	t.Run("join paths that resolve to current directory", func(t *testing.T) {
		base := New(0, []string{"a", "b"}, false)
		other := New(2, nil, false)    // ../..
		expected := New(0, nil, false) // a/b/../.. -> "" (contained)
		check(t, testScenario{
			name:              "a/b + ../..",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: expected,
		})
	})

	t.Run("join paths with leading dots", func(t *testing.T) {
		base := New(0, []string{"a"}, false)
		other := New(1, []string{"b"}, false)    // ../b
		expected := New(0, []string{"b"}, false) // a/../b -> b
		check(t, testScenario{
			name:              "a + ../b",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: expected,
		})
	})

	t.Run("join paths resulting in escaped path", func(t *testing.T) {
		base := New(0, []string{"a"}, false)
		other1 := New(2, nil, false) // ../..
		other2 := New(0, []string{"b"}, false)
		expected := New(1, []string{"b"}, false) // a/../../b -> ../b
		check(t, testScenario{
			name:              "a + ../.. + b",
			basePath:          base,
			others:            []*Path{other1, other2},
			expectedResult:    expected,
			expectedContained: nil,
		})
	})

	t.Run("join with base ending slash and other no slash", func(t *testing.T) {
		base := New(0, []string{"a"}, true) // a/
		other := New(0, []string{"b"}, false)
		expected := New(0, []string{"a", "b"}, false) // a/b (path.Join strips original slash if not from last part)
		check(t, testScenario{
			name:              "a/ + b",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: expected,
		})
	})

	t.Run("join with base no slash and other ending slash", func(t *testing.T) {
		base := New(0, []string{"a"}, false)
		other := New(0, []string{"b"}, true)         // b/
		expected := New(0, []string{"a", "b"}, true) // a/b/
		check(t, testScenario{
			name:              "a + b/",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: expected,
		})
	})

	t.Run("join with empty 'others' slice", func(t *testing.T) {
		base := New(0, []string{"a", "b"}, false)
		expected := New(0, []string{"a", "b"}, false)
		check(t, testScenario{
			name:              "a/b + (empty others)",
			basePath:          base,
			others:            []*Path{},
			expectedResult:    expected,
			expectedContained: expected,
		})
	})

	t.Run("join with empty string base path", func(t *testing.T) {
		base := New(0, nil, false) // ""
		other := New(0, []string{"a"}, false)
		expected := New(0, []string{"a"}, false) // "a"
		check(t, testScenario{
			name:              "\"\" + \"a\"",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: expected,
		})
	})

	t.Run("join with empty string other path", func(t *testing.T) {
		base := New(0, []string{"a"}, false)
		other := New(0, nil, false)              // ""
		expected := New(0, []string{"a"}, false) // "a"
		check(t, testScenario{
			name:              "\"a\" + \"\"",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: expected,
		})
	})

	t.Run("join root path with relative path", func(t *testing.T) {
		base := New(-1, nil, false) // "/"
		other := New(0, []string{"a", "b"}, false)
		expected := New(-1, []string{"a", "b"}, false) // "/a/b"
		check(t, testScenario{
			name:              "/ + a/b",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: nil,
		})
	})

	t.Run("join relative path with root path", func(t *testing.T) {
		base := New(0, []string{"a", "b"}, false)
		other := New(-1, nil, false)    // "/"
		expected := New(-1, nil, false) // "/" (root takes precedence)
		check(t, testScenario{
			name:              "a/b + /",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: nil,
		})
	})

	t.Run("join escaped path that becomes absolute", func(t *testing.T) {
		// e.g., /foo/bar.Join(../../..).Join(baz) => /baz
		base, _ := Parse("/foo/bar")
		other1, _ := Parse("../../..")
		other2, _ := Parse("baz")

		contained, result := base.Join(other1, other2)
		expected, _ := Parse("/baz")

		require.Equal(t, expected, result)
		require.Nil(t, contained)
	})

	t.Run("join escaped path with relative path that navigates further up", func(t *testing.T) {
		base, _ := Parse("a/b")
		other, _ := Parse("../../c") // a/b/../../c -> c
		contained, result := base.Join(other)
		expected, _ := Parse("c")

		require.Equal(t, expected, result)
		require.NotNil(t, contained)
		require.True(t, contained == result)
	})

	t.Run("join multiple escaped paths that result in contained path", func(t *testing.T) {
		base, _ := Parse("a/b") // Path(parents=0, parts={"a", "b"})
		// "a/b/../c/../d" -> "a/d"
		contained, result := base.Join(New(1, []string{"c"}, false), New(1, []string{"d"}, false))
		expected, _ := Parse("a/d")
		require.Equal(t, expected, result)
		require.NotNil(t, contained)
		require.True(t, contained == result)
	})

	t.Run("join absolute path with relative path ending with slash", func(t *testing.T) {
		base := New(-1, []string{"a", "b"}, false)              // "/a/b"
		other := New(0, []string{"c", "d"}, true)               // "c/d/"
		expected := New(-1, []string{"a", "b", "c", "d"}, true) // "/a/b/c/d/"
		check(t, testScenario{
			name:              "/a/b + c/d/",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: nil,
		})
	})

	t.Run("join relative path ending with slash and another relative path", func(t *testing.T) {
		base := New(0, []string{"a", "b"}, true)                // "a/b/"
		other := New(0, []string{"c", "d"}, false)              // "c/d"
		expected := New(0, []string{"a", "b", "c", "d"}, false) // "a/b/c/d"
		check(t, testScenario{
			name:              "a/b/ + c/d",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: expected,
		})
	})

	t.Run("join with a nil path should behave correctly (skip)", func(t *testing.T) {
		base := New(0, []string{"a"}, false)
		other1 := New(0, []string{"b"}, false)
		other2 := (*Path)(nil) // nil path
		other3 := New(0, []string{"c"}, false)
		expected := New(0, []string{"a", "b", "c"}, false)
		check(t, testScenario{
			name:              "a + b + nil + c",
			basePath:          base,
			others:            []*Path{other1, other2, other3},
			expectedResult:    expected,
			expectedContained: expected,
		})
	})

	t.Run("join with only nil paths", func(t *testing.T) {
		base := New(0, []string{"a"}, false)
		other1 := (*Path)(nil)
		other2 := (*Path)(nil)
		expected := New(0, []string{"a"}, false) // Should return original path
		check(t, testScenario{
			name:              "a + nil + nil",
			basePath:          base,
			others:            []*Path{other1, other2},
			expectedResult:    expected,
			expectedContained: expected,
		})
	})

	t.Run("join with an empty string as input (treated as '.')", func(t *testing.T) {
		base := New(0, []string{"foo"}, false)
		other := New(0, nil, false) // Represents "" which Parse treats as "."
		expected := New(0, []string{"foo"}, false)
		check(t, testScenario{
			name:              "foo + \"\" (empty path)",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: expected,
		})
	})

	t.Run("join an empty path with an empty path", func(t *testing.T) {
		base := New(0, nil, false)     // ""
		other := New(0, nil, false)    // ""
		expected := New(0, nil, false) // ""
		check(t, testScenario{
			name:              "\"\" + \"\" (empty paths)",
			basePath:          base,
			others:            []*Path{other},
			expectedResult:    expected,
			expectedContained: expected,
		})
	})

}
