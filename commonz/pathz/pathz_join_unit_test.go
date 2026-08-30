package pathz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitPathJoin(t *testing.T) {
	type testScenario struct {
		name                        string
		basePath                    *Path
		others                      []*Path
		expectedResult              *Path
		expectContainedReferenceEqual bool
	}

	tests := []testScenario{
		{
			name:                        "join two relative paths",
			basePath:                    New(0, []string{"a", "b"}, false),
			others:                      []*Path{New(0, []string{"c", "d"}, false)},
			expectedResult:              New(0, []string{"a", "b", "c", "d"}, false),
			expectContainedReferenceEqual: true,
		},
		{
			name:                        "join relative path with absolute path (absolute takes precedence)",
			basePath:                    New(0, []string{"a", "b"}, false),
			others:                      []*Path{New(-1, []string{"c", "d"}, false)},
			expectedResult:              New(-1, []string{"c", "d"}, false),
			expectContainedReferenceEqual: false, // Absolute path is not contained
		},
		{
			name:                        "join absolute path with relative path",
			basePath:                    New(-1, []string{"a", "b"}, false),
			others:                      []*Path{New(0, []string{"c", "d"}, false)},
			expectedResult:              New(-1, []string{"a", "b", "c", "d"}, false),
			expectContainedReferenceEqual: false,
		},
		{
			name:                        "join relative path with escaped path",
			basePath:                    New(0, []string{"a", "b"}, false),
			others:                      []*Path{New(1, []string{"c", "d"}, false)}, // ../c/d
			expectedResult:              New(0, []string{"a", "c", "d"}, false),      // a/b/../c/d -> a/c/d
			expectContainedReferenceEqual: true,
		},
		{
			name:                        "join escaped path with relative path",
			basePath:                    New(1, []string{"a", "b"}, false), // ../a/b
			others:                      []*Path{New(0, []string{"c", "d"}, false)},
			expectedResult:              New(1, []string{"a", "b", "c", "d"}, false), // ../a/b/c/d
			expectContainedReferenceEqual: false,                                      // Escaped path is not contained
		},
		{
			name:                        "join with empty base path",
			basePath:                    New(0, nil, false), // ""
			others:                      []*Path{New(0, []string{"a", "b"}, false)},
			expectedResult:              New(0, []string{"a", "b"}, false),
			expectContainedReferenceEqual: true,
		},
		{
			name:                        "join with multiple others",
			basePath:                    New(0, []string{"a"}, false),
			others:                      []*Path{New(0, []string{"b"}, false), New(0, []string{"c"}, false)},
			expectedResult:              New(0, []string{"a", "b", "c"}, false),
			expectContainedReferenceEqual: true,
		},
		{
			name:                        "join with intermediate absolute path",
			basePath:                    New(0, []string{"a"}, false),
			others:                      []*Path{New(0, []string{"b"}, false), New(-1, []string{"c"}, false), New(0, []string{"d"}, false)},
			expectedResult:              New(-1, []string{"c", "d"}, false), // /c overrides a/b, then d joins
			expectContainedReferenceEqual: false,
		},
		{
			name:                        "join paths that resolve to current directory",
			basePath:                    New(0, []string{"a", "b"}, false),
			others:                      []*Path{New(2, nil, false)},    // ../..
			expectedResult:              New(0, nil, false),             // a/b/../.. -> "" (contained)
			expectContainedReferenceEqual: true,
		},
		{
			name:                        "join paths with leading dots",
			basePath:                    New(0, []string{"a"}, false),
			others:                      []*Path{New(1, []string{"b"}, false)}, // ../b
			expectedResult:              New(0, []string{"b"}, false),          // a/../b -> b
			expectContainedReferenceEqual: true,
		},
		{
			name:                        "join paths resulting in escaped path",
			basePath:                    New(0, []string{"a"}, false),
			others:                      []*Path{New(2, nil, false), New(0, []string{"b"}, false)}, // ../.., b
			expectedResult:              New(1, []string{"b"}, false),                              // a/../../b -> ../b
			expectContainedReferenceEqual: false,
		},
		{
			name:                        "join with base ending slash and other no slash",
			basePath:                    New(0, []string{"a"}, true), // a/
			others:                      []*Path{New(0, []string{"b"}, false)},
			expectedResult:              New(0, []string{"a", "b"}, false), // a/b (path.Join strips original slash if not from last part)
			expectContainedReferenceEqual: true,
		},
		{
			name:                        "join with base no slash and other ending slash",
			basePath:                    New(0, []string{"a"}, false),
			others:                      []*Path{New(0, []string{"b"}, true)}, // b/
			expectedResult:              New(0, []string{"a", "b"}, true),     // a/b/
			expectContainedReferenceEqual: true,
		},
		{
			name:                        "join with empty 'others' slice",
			basePath:                    New(0, []string{"a", "b"}, false),
			others:                      []*Path{},
			expectedResult:              New(0, []string{"a", "b"}, false),
			expectContainedReferenceEqual: true,
		},
		{
			name:                        "join with empty string base path",
			basePath:                    New(0, nil, false), // ""
			others:                      []*Path{New(0, []string{"a"}, false)},
			expectedResult:              New(0, []string{"a"}, false), // "a"
			expectContainedReferenceEqual: true,
		},
		{
			name:                        "join with empty string other path",
			basePath:                    New(0, []string{"a"}, false),
			others:                      []*Path{New(0, nil, false)},              // ""
			expectedResult:              New(0, []string{"a"}, false),             // "a"
			expectContainedReferenceEqual: true,
		},
		{
			name:                        "join root path with relative path",
			basePath:                    New(-1, nil, false), // "/"
			others:                      []*Path{New(0, []string{"a", "b"}, false)},
			expectedResult:              New(-1, []string{"a", "b"}, false), // "/a/b"
			expectContainedReferenceEqual: false,
		},
		{
			name:                        "join relative path with root path",
			basePath:                    New(0, []string{"a", "b"}, false),
			others:                      []*Path{New(-1, nil, false)},    // "/"
			expectedResult:              New(-1, nil, false),             // "/" (root takes precedence)
			expectContainedReferenceEqual: false,
		},
		{
			name: "join escaped path that becomes absolute",
			// e.g., /foo/bar.Join(../../..).Join(baz) => /baz
			basePath:                    func() *Path { p, _ := Parse("/foo/bar"); return p }(),
			others:                      []*Path{func() *Path { p, _ := Parse("../../.."); return p }(), func() *Path { p, _ := Parse("baz"); return p }()},
			expectedResult:              func() *Path { p, _ := Parse("/baz"); return p }(),
			expectContainedReferenceEqual: false,
		},
		{
			name: "join escaped path with relative path that navigates further up",
			basePath:                    func() *Path { p, _ := Parse("a/b"); return p }(),
			others:                      []*Path{func() *Path { p, _ := Parse("../../c"); return p }()}, // a/b/../../c -> c
			expectedResult:              func() *Path { p, _ := Parse("c"); return p }(),
			expectContainedReferenceEqual: true,
		},
		{
			name: "join multiple escaped paths that result in contained path",
			basePath:                    func() *Path { p, _ := Parse("a/b"); return p }(), // Path(parents=0, parts={"a", "b"})
			// "a/b/../c/../d" -> "a/d"
			others:                      []*Path{New(1, []string{"c"}, false), New(1, []string{"d"}, false)},
			expectedResult:              func() *Path { p, _ := Parse("a/d"); return p }(),
			expectContainedReferenceEqual: true,
		},
		{
			name:                        "join absolute path with relative path ending with slash",
			basePath:                    New(-1, []string{"a", "b"}, false),              // "/a/b"
			others:                      []*Path{New(0, []string{"c", "d"}, true)},        // "c/d/"
			expectedResult:              New(-1, []string{"a", "b", "c", "d"}, true), // "/a/b/c/d/"
			expectContainedReferenceEqual: false,
		},
		{
			name:                        "join relative path ending with slash and another relative path",
			basePath:                    New(0, []string{"a", "b"}, true),                // "a/b/"
			others:                      []*Path{New(0, []string{"c", "d"}, false)},       // "c/d"
			expectedResult:              New(0, []string{"a", "b", "c", "d"}, false), // "a/b/c/d"
			expectContainedReferenceEqual: true,
		},
		{
			name:                        "join with a nil path should behave correctly (skip)",
			basePath:                    New(0, []string{"a"}, false),
			others:                      []*Path{New(0, []string{"b"}, false), (*Path)(nil), New(0, []string{"c"}, false)},
			expectedResult:              New(0, []string{"a", "b", "c"}, false),
			expectContainedReferenceEqual: true,
		},
		{
			name:                        "join with only nil paths",
			basePath:                    New(0, []string{"a"}, false),
			others:                      []*Path{(*Path)(nil), (*Path)(nil)},
			expectedResult:              New(0, []string{"a"}, false), // Should return original path
			expectContainedReferenceEqual: true,
		},
		{
			name:                        "join with an empty string as input (treated as '.')",
			basePath:                    New(0, []string{"foo"}, false),
			others:                      []*Path{New(0, nil, false)}, // Represents "" which Parse treats as "."
			expectedResult:              New(0, []string{"foo"}, false),
			expectContainedReferenceEqual: true,
		},
		{
			name:                        "join an empty path with an empty path",
			basePath:                    New(0, nil, false),     // ""
			others:                      []*Path{New(0, nil, false)},    // ""
			expectedResult:              New(0, nil, false), // ""
			expectContainedReferenceEqual: true,
		},
	}

	for _, s := range tests {
		s := s // Capture range variable
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()

			contained, result := s.basePath.Join(s.others...)

			require.Equal(t, s.expectedResult, result, "Result path mismatch for scenario: %s", s.name)

			if s.expectContainedReferenceEqual {
				require.NotNil(t, contained, "Contained path should not be nil for scenario: %s", s.name)
				// This checks for reference equality as per the requirement
				require.True(t, contained == result, "Contained path should be the same reference as result for scenario: %s", s.name)
				require.Equal(t, s.expectedResult, contained, "Contained path value mismatch for scenario: %s", s.name)
			} else {
				require.Nil(t, contained, "Contained path should be nil for scenario: %s", s.name)
			}
		})
	}
}


