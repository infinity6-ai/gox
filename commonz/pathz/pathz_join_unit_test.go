package pathz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitJoin(t *testing.T) {
	type testScenario struct {
		name          string
		basePathStr   string
		otherPathStrs []string
		expectedPath  *Path
		expectErr     bool
		errContains   string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		basePath, err := Parse(s.basePathStr)
		require.NoError(t, err, "failed to parse base path %q", s.basePathStr)

		others := make([]*Path, len(s.otherPathStrs))
		for i, pStr := range s.otherPathStrs {
			otherPath, err := Parse(pStr)
			require.NoError(t, err, "failed to parse other path %q at index %d", pStr, i)
			others[i] = otherPath
		}

		joinedPath, err := basePath.Join(others...)

		if s.expectErr {
			require.Error(t, err, "expected an error for scenario %q", s.name)
			require.ErrorIs(t, err, ErrEscaped, "expected ErrEscaped for scenario %q", s.name)
			require.Contains(t, err.Error(), s.errContains, "error message mismatch for scenario %q", s.name)
			if s.expectedPath != nil {
				require.Equal(t, s.expectedPath, joinedPath, "escaped path mismatch for scenario %q", s.name)
			}
			return
		}

		require.NoError(t, err, "did not expect an error for scenario %q", s.name)
		require.NotNil(t, joinedPath, "joined path should not be nil for scenario %q", s.name)
		require.Equal(t, s.expectedPath, joinedPath, "joined path mismatch for scenario %q", s.name)
	}

	t.Run("simple join", func(t *testing.T) {
		check(t, testScenario{
			name:          "a/b + c = a/b/c",
			basePathStr:   "a/b",
			otherPathStrs: []string{"c"},
			expectedPath:  New(0, []string{"a", "b", "c"}, false),
		})
	})

	t.Run("multiple simple joins", func(t *testing.T) {
		check(t, testScenario{
			name:          "a/b + c + d = a/b/c/d",
			basePathStr:   "a/b",
			otherPathStrs: []string{"c", "d"},
			expectedPath:  New(0, []string{"a", "b", "c", "d"}, false),
		})
	})

	t.Run("join with parent directory navigation staying within base", func(t *testing.T) {
		check(t, testScenario{
			name:          "a/b/c + ../d = a/b/d",
			basePathStr:   "a/b/c",
			otherPathStrs: []string{"../d"},
			expectedPath:  New(0, []string{"a", "b", "d"}, false),
			expectErr:     true,
			errContains:   "path escaped error: joining 'a/b/c' to '[../d]' results in 'a/b/d' which is outside the base",
		})
	})

	t.Run("join with multiple parent directory navigation staying within base", func(t *testing.T) {
		check(t, testScenario{
			name:          "a/b/c/d + ../../e = a/b/e",
			basePathStr:   "a/b/c/d",
			otherPathStrs: []string{"../../e"},
			expectedPath:  New(0, []string{"a", "b", "e"}, false),
			expectErr:     true,
			errContains:   "path escaped error: joining 'a/b/c/d' to '[../../e]' results in 'a/b/e' which is outside the base",
		})
	})

	t.Run("join with parent directory navigation escaping base", func(t *testing.T) {
		check(t, testScenario{
			name:          "a/b + ../../c = escaped",
			basePathStr:   "a/b",
			otherPathStrs: []string{"../../c"},
			expectErr:     true,
			errContains:   "path escaped error: joining 'a/b' to '[../../c]' results in 'c' which is outside the base",
			expectedPath:  New(0, []string{"c"}, false), // The resulting path after cleaning
		})
	})

	t.Run("join with absolute path as other (relative base)", func(t *testing.T) {
		check(t, testScenario{
			name:          "a/b + /c/d = escaped (absolute in relative)",
			basePathStr:   "a/b",
			otherPathStrs: []string{"/c/d"},
			expectErr:     true,
			errContains:   "path escaped error: joining 'a/b' to '[/c/d]' results in '/c/d' which is outside the base",
			expectedPath:  New(-1, []string{"c", "d"}, false),
		})
	})

	t.Run("join with absolute path as other (absolute base but different root)", func(t *testing.T) {
		check(t, testScenario{
			name:          "/a/b + /c/d = escaped (absolute in absolute, different root)",
			basePathStr:   "/a/b",
			otherPathStrs: []string{"/c/d"},
			expectErr:     true,
			errContains:   "path escaped error: joining '/a/b' to '[/c/d]' results in '/c/d' which is outside the base",
			expectedPath:  New(-1, []string{"c", "d"}, false),
		})
	})

	t.Run("join with absolute path as other (absolute base and valid descendant)", func(t *testing.T) {
		check(t, testScenario{
			name:          "/a/b + /a/b/c/d = valid",
			basePathStr:   "/a/b",
			otherPathStrs: []string{"/a/b/c/d"},
			expectErr:     false,
			expectedPath:  New(-1, []string{"a", "b", "c", "d"}, false),
		})
	})

	t.Run("join with empty other path", func(t *testing.T) {
		check(t, testScenario{
			name:          "a/b + '' = a/b",
			basePathStr:   "a/b",
			otherPathStrs: []string{""},
			expectedPath:  New(0, []string{"a", "b"}, false),
		})
	})

	t.Run("join with current directory .", func(t *testing.T) {
		check(t, testScenario{
			name:          "a/b + . = a/b",
			basePathStr:   "a/b",
			otherPathStrs: []string{"."},
			expectedPath:  New(0, []string{"a", "b"}, false),
		})
	})

	t.Run("join with current directory ./ and trailing slash", func(t *testing.T) {
		check(t, testScenario{
			name:          "a/b + ./ = a/b/",
			basePathStr:   "a/b",
			otherPathStrs: []string{"./"},
			expectedPath:  New(0, []string{"a", "b"}, true),
		})
	})

	t.Run("base path with trailing slash, simple join", func(t *testing.T) {
		check(t, testScenario{
			name:          "a/b/ + c = a/b/c",
			basePathStr:   "a/b/",
			otherPathStrs: []string{"c"},
			expectedPath:  New(0, []string{"a", "b", "c"}, false),
		})
	})

	t.Run("base path with trailing slash, other path with trailing slash", func(t *testing.T) {
		check(t, testScenario{
			name:          "a/b/ + c/ = a/b/c/",
			basePathStr:   "a/b/",
			otherPathStrs: []string{"c/"},
			expectedPath:  New(0, []string{"a", "b", "c"}, true),
		})
	})

	t.Run("join with leading ../ (base is relative)", func(t *testing.T) {
		check(t, testScenario{
			name:          "b/c + ../../a = escaped",
			basePathStr:   "b/c",
			otherPathStrs: []string{"../../a"},
			expectErr:     true,
			errContains:   "path escaped error: joining '../../a' to 'b/c' results in '../a' which is outside the base",
			expectedPath:  New(1, []string{"a"}, false),
		})
	})

	t.Run("join with leading ../ (base is absolute)", func(t *testing.T) {
		check(t, testScenario{
			name:          "/b/c + ../../a = escaped",
			basePathStr:   "/b/c",
			otherPathStrs: []string{"../../a"},
			expectErr:     true,
			errContains:   "path escaped error: joining '../../a' to '/b/c' results in '/a' which is outside the base",
			expectedPath:  New(-1, []string{"a"}, false),
		})
	})

	t.Run("join root with a relative path", func(t *testing.T) {
		check(t, testScenario{
			name:          "/ + a/b = /a/b",
			basePathStr:   "/",
			otherPathStrs: []string{"a/b"},
			expectedPath:  New(-1, []string{"a", "b"}, false),
		})
	})

	t.Run("join root with an absolute path", func(t *testing.T) {
		check(t, testScenario{
			name:          "/ + /a/b = /a/b",
			basePathStr:   "/",
			otherPathStrs: []string{"/a/b"},
			expectedPath:  New(-1, []string{"a", "b"}, false),
		})
	})

	t.Run("join with no other paths", func(t *testing.T) {
		check(t, testScenario{
			name:          "a/b + (no others) = a/b",
			basePathStr:   "a/b",
			otherPathStrs: []string{},
			expectedPath:  New(0, []string{"a", "b"}, false),
		})
	})

	t.Run("path a/b/../c", func(t *testing.T) {
		check(t, testScenario{
			name:          "path a/b/../c",
			basePathStr:   "a",
			otherPathStrs: []string{"b/../c"},
			expectedPath:  New(0, []string{"a", "c"}, false),
		})
	})
	t.Run("absolute path join with .. and ending slash", func(t *testing.T) {
		check(t, testScenario{
			name:          "absolute path join with .. and ending slash",
			basePathStr:   "/a/b",
			otherPathStrs: []string{"../c/"},
			expectedPath:  New(-1, []string{"a", "c"}, true),
		})
	})

	t.Run("path ../a/b join with ../../c escaping", func(t *testing.T) {
		check(t, testScenario{
			name:          "path ../a/b join with ../../c escaping",
			basePathStr:   "../a/b",
			otherPathStrs: []string{"../../c"},
			expectErr:     true,
			errContains:   "path escaped error: joining '../a/b' to '[../../c]' results in '../c' which is outside the base",
			expectedPath:  New(1, []string{"c"}, false),
		})
	})
}
