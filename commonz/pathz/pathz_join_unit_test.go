package pathz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitPathJoin(t *testing.T) {
	type testScenario struct {
		basePath                      *Path
		others                        []*Path
		expectedResult                *Path
		expectContainedReferenceEqual bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		contained, result := s.basePath.Join(s.others...)

		require.Equal(t, s.expectedResult, result, "Result path mismatch")

		if s.expectContainedReferenceEqual {
			require.NotNil(t, contained, "Contained path should not be nil")
			// This checks for reference equality as per the requirement
			require.True(t, contained == result, "Contained path should be the same reference as result")
			require.Equal(t, s.expectedResult, contained, "Contained path value mismatch")
		} else {
			require.Nil(t, contained, "Contained path should be nil")
		}
	}

	t.Run("join two relative paths", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"a", "b"}, false),
			others:                        []*Path{New(0, []string{"c", "d"}, false)},
			expectedResult:                New(0, []string{"a", "b", "c", "d"}, false),
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join relative path with absolute path (absolute takes precedence)", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"a", "b"}, false),
			others:                        []*Path{New(-1, []string{"c", "d"}, false)},
			expectedResult:                New(-1, []string{"c", "d"}, false),
			expectContainedReferenceEqual: false, // Absolute path is not contained
		})
	})

	t.Run("join absolute path with relative path", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(-1, []string{"a", "b"}, false),
			others:                        []*Path{New(0, []string{"c", "d"}, false)},
			expectedResult:                New(-1, []string{"a", "b", "c", "d"}, false),
			expectContainedReferenceEqual: false,
		})
	})

	t.Run("join relative path with escaped path", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"a", "b"}, false),
			others:                        []*Path{New(1, []string{"c", "d"}, false)}, // ../c/d
			expectedResult:                New(0, []string{"a", "c", "d"}, false),     // a/b/../c/d -> a/c/d
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join escaped path with relative path", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(1, []string{"a", "b"}, false), // ../a/b
			others:                        []*Path{New(0, []string{"c", "d"}, false)},
			expectedResult:                New(1, []string{"a", "b", "c", "d"}, false), // ../a/b/c/d
			expectContainedReferenceEqual: false,                                       // Escaped path is not contained
		})
	})

	t.Run("join with empty base path", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, nil, false), // ""
			others:                        []*Path{New(0, []string{"a", "b"}, false)},
			expectedResult:                New(0, []string{"a", "b"}, false),
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join with multiple others", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"a"}, false),
			others:                        []*Path{New(0, []string{"b"}, false), New(0, []string{"c"}, false)},
			expectedResult:                New(0, []string{"a", "b", "c"}, false),
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join with intermediate absolute path", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"a"}, false),
			others:                        []*Path{New(0, []string{"b"}, false), New(-1, []string{"c"}, false), New(0, []string{"d"}, false)},
			expectedResult:                New(-1, []string{"c", "d"}, false), // /c overrides a/b, then d joins
			expectContainedReferenceEqual: false,
		})
	})

	t.Run("join paths that resolve to current directory", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"a", "b"}, false),
			others:                        []*Path{New(2, nil, false)}, // ../..
			expectedResult:                New(0, nil, false),          // a/b/../.. -> "" (contained)
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join paths with leading dots", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"a"}, false),
			others:                        []*Path{New(1, []string{"b"}, false)}, // ../b
			expectedResult:                New(0, []string{"b"}, false),          // a/../b -> b
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join paths resulting in escaped path", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"a"}, false),
			others:                        []*Path{New(2, nil, false), New(0, []string{"b"}, false)}, // ../.., b
			expectedResult:                New(1, []string{"b"}, false),                              // a/../../b -> ../b
			expectContainedReferenceEqual: false,
		})
	})

	t.Run("join with base ending slash and other no slash", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"a"}, true), // a/
			others:                        []*Path{New(0, []string{"b"}, false)},
			expectedResult:                New(0, []string{"a", "b"}, false), // a/b (path.Join strips original slash if not from last part)
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join with base no slash and other ending slash", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"a"}, false),
			others:                        []*Path{New(0, []string{"b"}, true)}, // b/
			expectedResult:                New(0, []string{"a", "b"}, true),     // a/b/
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join with empty 'others' slice", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"a", "b"}, false),
			others:                        []*Path{},
			expectedResult:                New(0, []string{"a", "b"}, false),
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join with empty string base path", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, nil, false), // ""
			others:                        []*Path{New(0, []string{"a"}, false)},
			expectedResult:                New(0, []string{"a"}, false), // "a"
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join with empty string other path", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"a"}, false),
			others:                        []*Path{New(0, nil, false)},  // ""
			expectedResult:                New(0, []string{"a"}, false), // "a"
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join root path with relative path", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(-1, nil, false), // "/"
			others:                        []*Path{New(0, []string{"a", "b"}, false)},
			expectedResult:                New(-1, []string{"a", "b"}, false), // "/a/b"
			expectContainedReferenceEqual: false,
		})
	})

	t.Run("join relative path with root path", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"a", "b"}, false),
			others:                        []*Path{New(-1, nil, false)}, // "/"
			expectedResult:                New(-1, nil, false),          // "/" (root takes precedence)
			expectContainedReferenceEqual: false,
		})
	})

	t.Run("join escaped path that becomes absolute", func(t *testing.T) {
		check(t, testScenario{
			// e.g., /foo/bar.Join(../../..).Join(baz) => /baz
			basePath:                      func() *Path { p, _ := Parse("/foo/bar"); return p }(),
			others:                        []*Path{func() *Path { p, _ := Parse("../../.."); return p }(), func() *Path { p, _ := Parse("baz"); return p }()},
			expectedResult:                func() *Path { p, _ := Parse("/baz"); return p }(),
			expectContainedReferenceEqual: false,
		})
	})

	t.Run("join escaped path with relative path that navigates further up", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      func() *Path { p, _ := Parse("a/b"); return p }(),
			others:                        []*Path{func() *Path { p, _ := Parse("../../c"); return p }()}, // a/b/../../c -> c
			expectedResult:                func() *Path { p, _ := Parse("c"); return p }(),
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join multiple escaped paths that result in contained path", func(t *testing.T) {
		check(t, testScenario{
			basePath: func() *Path { p, _ := Parse("a/b"); return p }(), // Path(parents=0, parts={"a", "b"})
			// "a/b/../c/../d" -> "a/d"
			others:                        []*Path{New(1, []string{"c"}, false), New(1, []string{"d"}, false)},
			expectedResult:                func() *Path { p, _ := Parse("a/d"); return p }(),
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join absolute path with relative path ending with slash", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(-1, []string{"a", "b"}, false),          // "/a/b"
			others:                        []*Path{New(0, []string{"c", "d"}, true)},   // "c/d/"
			expectedResult:                New(-1, []string{"a", "b", "c", "d"}, true), // "/a/b/c/d/"
			expectContainedReferenceEqual: false,
		})
	})

	t.Run("join relative path ending with slash and another relative path", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"a", "b"}, true),            // "a/b/"
			others:                        []*Path{New(0, []string{"c", "d"}, false)},  // "c/d"
			expectedResult:                New(0, []string{"a", "b", "c", "d"}, false), // "a/b/c/d"
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join with a nil path should behave correctly (skip)", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"a"}, false),
			others:                        []*Path{New(0, []string{"b"}, false), (*Path)(nil), New(0, []string{"c"}, false)},
			expectedResult:                New(0, []string{"a", "b", "c"}, false),
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join with only nil paths", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"a"}, false),
			others:                        []*Path{(*Path)(nil), (*Path)(nil)},
			expectedResult:                New(0, []string{"a"}, false), // Should return original path
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join with an empty string as input (treated as '.')", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, []string{"foo"}, false),
			others:                        []*Path{New(0, nil, false)}, // Represents "" which Parse treats as "."
			expectedResult:                New(0, []string{"foo"}, false),
			expectContainedReferenceEqual: true,
		})
	})

	t.Run("join an empty path with an empty path", func(t *testing.T) {
		check(t, testScenario{
			basePath:                      New(0, nil, false),          // ""
			others:                        []*Path{New(0, nil, false)}, // ""
			expectedResult:                New(0, nil, false),          // ""
			expectContainedReferenceEqual: true,
		})
	})
}
