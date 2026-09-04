package pathz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/stretchr/testify/require"
)

func TestUnitParsePattern(t *testing.T) {
	type testScenario struct {
		name          string
		pathStr       string
		patternStr    string
		expectedParams map[string]string
		expectErr     bool
		errContains   string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		p, err := pathz.Parse(s.pathStr)
		require.NoError(t, err, "failed to parse path %q", s.pathStr)
		pattern, err := pathz.Parse(s.patternStr)
		require.NoError(t, err, "failed to parse pattern %q", s.patternStr)

		params, err := p.ParsePattern(pattern)

		if s.expectErr {
			require.Error(t, err, "expected error for scenario %q", s.name)
			require.Contains(t, err.Error(), s.errContains, "error message mismatch for scenario %q", s.name)
			require.Nil(t, params)
		} else {
			require.NoError(t, err, "did not expect error for scenario %q", s.name)
			require.Equal(t, s.expectedParams, params, "params mismatch for scenario %q", s.name)
		}
	}

	t.Run("simple pattern match", func(t *testing.T) {
		check(t, testScenario{
			name:          "simple pattern match",
			pathStr:       "a/value1/b/value2",
			patternStr:    "a/{param1}/b/{param2}",
			expectedParams: map[string]string{"param1": "value1", "param2": "value2"},
		})
	})

	t.Run("pattern with no parameters", func(t *testing.T) {
		check(t, testScenario{
			name:          "pattern with no parameters",
			pathStr:       "a/b/c",
			patternStr:    "a/b/c",
			expectedParams: map[string]string{},
		})
	})

	t.Run("pattern with single parameter", func(t *testing.T) {
		check(t, testScenario{
			name:          "pattern with single parameter",
			pathStr:       "users/123",
			patternStr:    "users/{id}",
			expectedParams: map[string]string{"id": "123"},
		})
	})

	t.Run("absolute paths", func(t *testing.T) {
		check(t, testScenario{
			name:          "absolute paths",
			pathStr:       "/api/v1/resource",
			patternStr:    "/api/{version}/resource",
			expectedParams: map[string]string{"version": "v1"},
		})
	})

	t.Run("mismatch - literal part", func(t *testing.T) {
		check(t, testScenario{
			name:        "mismatch - literal part",
			pathStr:     "a/value1/x/value2",
			patternStr:  "a/{param1}/b/{param2}",
			expectErr:   true,
			errContains: "literal part mismatch at index 2: expected 'b', got 'x'",
		})
	})

	t.Run("mismatch - path length", func(t *testing.T) {
		check(t, testScenario{
			name:        "mismatch - path length",
			pathStr:     "a/value1/b",
			patternStr:  "a/{param1}/b/{param2}",
			expectErr:   true,
			errContains: "path length mismatch: path has 3 parts, pattern has 4 parts",
		})
	})

	t.Run("mismatch - path parents (absolute vs relative)", func(t *testing.T) {
		check(t, testScenario{
			name:        "mismatch - path parents (absolute vs relative)",
			pathStr:     "/a/b",
			patternStr:  "a/b",
			expectErr:   true,
			errContains: "path parents mismatch: path parents -1, pattern parents 0",
		})
	})

	t.Run("mismatch - path parents (different relative levels)", func(t *testing.T) {
		check(t, testScenario{
			name:        "mismatch - path parents (different relative levels)",
			pathStr:     "../a/b",
			patternStr:  "../../a/b",
			expectErr:   true,
			errContains: "path parents mismatch: path parents 1, pattern parents 2",
		})
	})

	t.Run("path with ending slash", func(t *testing.T) {
		check(t, testScenario{
			name:          "path with ending slash",
			pathStr:       "a/value1/",
			patternStr:    "a/{param1}/",
			expectedParams: map[string]string{"param1": "value1"},
		})
	})

	t.Run("mismatch - ending slash", func(t *testing.T) {
		check(t, testScenario{
			name:        "mismatch - ending slash",
			pathStr:     "a/value1",
			patternStr:  "a/{param1}/",
			expectErr:   true,
			errContains: "path ending slash mismatch: path has ending slash false, pattern has ending slash true",
		})
	})

	t.Run("empty parameter name in pattern", func(t *testing.T) {
		check(t, testScenario{
			name:        "empty parameter name in pattern",
			pathStr:     "a/c/b",
			patternStr:  "a/{}/b",
			expectErr:   true,
			errContains: "empty parameter name in pattern part: '{}'",
		})
	})

	t.Run("duplicate parameter name in pattern", func(t *testing.T) {
		check(t, testScenario{
			name:        "duplicate parameter name in pattern",
			pathStr:     "a/v1/v2",
			patternStr:  "a/{param}/{param}",
			expectErr:   true,
			errContains: "duplicate parameter name found: 'param'",
		})
	})

	t.Run("illegal character in parameter name - slash", func(t *testing.T) {
		check(t, testScenario{
			name:        "illegal character in parameter name - slash",
			pathStr:     "a/v1",
			patternStr:  "a/{pa/ram}",
			expectErr:   true,
			errContains: "path length mismatch",
		})
	})
	t.Run("illegal character in parameter name - curly bracket", func(t *testing.T) {
		check(t, testScenario{
			name:        "illegal character in parameter name - curly bracket",
			pathStr:     "a/v1",
			patternStr:  "a/{pa{ram}",
			expectErr:   true,
			errContains: "illegal character in parameter name 'pa{ram'",
		})
	})
}

func TestUnitMustParsePattern(t *testing.T) {
	// Test case for successful parsing
	path, _ := pathz.Parse("a/value1/b/value2")
	pattern, _ := pathz.Parse("a/{param1}/b/{param2}")
	expectedParams := map[string]string{"param1": "value1", "param2": "value2"}
	params := path.MustParsePattern(pattern)
	require.Equal(t, expectedParams, params)

	// Test case for a panic due to error
	pathErr, _ := pathz.Parse("a/value1/x/value2")
	patternErr, _ := pathz.Parse("a/{param1}/b/{param2}")
	require.Panics(t, func() {
		pathErr.MustParsePattern(patternErr)
	})
}

func TestUnitFormatPattern(t *testing.T) {
	type testScenario struct {
		name          string
		patternStr    string
		params        map[string]string
		expectedPathStr string
		expectErr     bool
		errContains   string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		pattern, err := pathz.Parse(s.patternStr)
		require.NoError(t, err, "failed to parse pattern %q", s.patternStr)

		formattedPath, err := (&pathz.Path{}).FormatPattern(pattern, s.params) // Use an empty path as receiver, as it's not used by FormatPattern

		if s.expectErr {
			require.Error(t, err, "expected error for scenario %q", s.name)
			require.Contains(t, err.Error(), s.errContains, "error message mismatch for scenario %q", s.name)
			require.Nil(t, formattedPath)
		} else {
			require.NoError(t, err, "did not expect error for scenario %q", s.name)
			expectedPath, err := pathz.Parse(s.expectedPathStr)
			require.NoError(t, err, "failed to parse expected path %q", s.expectedPathStr)
			require.Equal(t, expectedPath, formattedPath, "formatted path mismatch for scenario %q", s.name)
		}
	}

	t.Run("simple format", func(t *testing.T) {
		check(t, testScenario{
			name:          "simple format",
			patternStr:    "a/{param1}/b/{param2}",
			params:        map[string]string{"param1": "value1", "param2": "value2"},
			expectedPathStr: "a/value1/b/value2",
		})
	})

	t.Run("format with no parameters", func(t *testing.T) {
		check(t, testScenario{
			name:          "format with no parameters",
			patternStr:    "a/b/c",
			params:        map[string]string{},
			expectedPathStr: "a/b/c",
		})
	})

	t.Run("format with single parameter", func(t *testing.T) {
		check(t, testScenario{
			name:          "format with single parameter",
			patternStr:    "users/{id}",
			params:        map[string]string{"id": "123"},
			expectedPathStr: "users/123",
		})
	})

	t.Run("absolute path format", func(t *testing.T) {
		check(t, testScenario{
			name:          "absolute path format",
			patternStr:    "/api/{version}/resource",
			params:        map[string]string{"version": "v1"},
			expectedPathStr: "/api/v1/resource",
		})
	})

	t.Run("missing parameter", func(t *testing.T) {
		check(t, testScenario{
			name:        "missing parameter",
			patternStr:  "a/{param1}/b/{param2}",
			params:        map[string]string{"param1": "value1"},
			expectErr:   true,
			errContains: "missing parameter value for 'param2' in pattern part index 3 ('{param2}')",
		})
	})

	t.Run("extra parameter", func(t *testing.T) {
		check(t, testScenario{
			name:        "extra parameter",
			patternStr:  "a/{param1}/b",
			params:        map[string]string{"param1": "value1", "extra": "value"},
			expectErr:   true,
			errContains: "extra parameter provided: 'extra' not found in pattern",
		})
	})

	t.Run("path with ending slash format", func(t *testing.T) {
		check(t, testScenario{
			name:          "path with ending slash format",
			patternStr:    "a/{param1}/",
			params:        map[string]string{"param1": "value1"},
			expectedPathStr: "a/value1/",
		})
	})

	t.Run("empty parameter name in pattern", func(t *testing.T) {
		check(t, testScenario{
			name:        "empty parameter name in pattern",
			patternStr:  "a/{}/b",
			params:        map[string]string{"": "val"}, // This param would be technically present, but name invalid
			expectErr:   true,
			errContains: "empty parameter name in pattern part: '{}'",
		})
	})

	t.Run("illegal character in parameter name - slash", func(t *testing.T) {
		check(t, testScenario{
			name:        "illegal character in parameter name - slash",
			patternStr:  "a/{pa/ram}",
			params:        map[string]string{"pa/ram": "v1"},
			expectErr:   true,
			errContains: "extra parameter provided: 'pa/ram' not found in pattern",
		})
	})
	t.Run("illegal character in parameter name - curly bracket", func(t *testing.T) {
		check(t, testScenario{
			name:        "illegal character in parameter name - curly bracket",
			patternStr:  "a/{pa{ram}",
			params:        map[string]string{"pa{ram": "v1"},
			expectErr:   true,
			errContains: "illegal character in parameter name 'pa{ram'",
		})
	})
}

func TestUnitMustFormatPattern(t *testing.T) {
	// Test case for successful formatting
	pattern, _ := pathz.Parse("a/{param1}/b/{param2}")
	params := map[string]string{"param1": "value1", "param2": "value2"}
	expectedPath, _ := pathz.Parse("a/value1/b/value2")
	formattedPath := (&pathz.Path{}).MustFormatPattern(pattern, params)
	require.Equal(t, expectedPath, formattedPath)

	// Test case for a panic due to error (e.g., missing parameter)
	patternErr, _ := pathz.Parse("a/{param1}/b/{param2}")
	paramsErr := map[string]string{"param1": "value1"}
	require.Panics(t, func() {
		(&pathz.Path{}).MustFormatPattern(patternErr, paramsErr)
	})
}

func TestUnitPathParseAndFormatPatternEndToEnd(t *testing.T) {
	// Define a pattern
	patternStr := "/users/{userID}/posts/{postID}"
	pattern, err := pathz.Parse(patternStr)
	require.NoError(t, err)

	// Define a concrete path that matches the pattern
	pathStr := "/users/123/posts/abc"
	p, err := pathz.Parse(pathStr)
	require.NoError(t, err)

	// 1. Test ParsePattern: Extract parameters from the path using the pattern
	extractedParams, err := p.ParsePattern(pattern)
	require.NoError(t, err)
	expectedParams := map[string]string{"userID": "123", "postID": "abc"}
	require.Equal(t, expectedParams, extractedParams)

	// 2. Test FormatPattern: Reconstruct the path using the pattern and extracted parameters
	formattedPath, err := (&pathz.Path{}).FormatPattern(pattern, extractedParams)
	require.NoError(t, err)
	require.Equal(t, p, formattedPath) // The formatted path should be equal to the original path

	// Test with a different set of parameters
	newParams := map[string]string{"userID": "456", "postID": "xyz"}
	expectedNewPathStr := "/users/456/posts/xyz"
	expectedNewPath, err := pathz.Parse(expectedNewPathStr)
	require.NoError(t, err)

	formattedNewPath, err := (&pathz.Path{}).FormatPattern(pattern, newParams)
	require.NoError(t, err)
	require.Equal(t, expectedNewPath, formattedNewPath)
}
