package patternpathz

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/stretchr/testify/require"
)

func TestUnitParse(t *testing.T) {
	type testScenario struct {
		name       string
		patternStr string
		expectErr  bool
		errMsg     string
		expectedSegments []string
		expectedNames map[string]int
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		patternPath, err := pathz.Parse(s.patternStr)
		require.NoError(t, err)

		pp, err := Parse(patternPath)

		if s.expectErr {
			require.Error(t, err, "Expected an error for scenario: %s", s.name)
			require.Contains(t, err.Error(), s.errMsg, "Error message mismatch for scenario: %s", s.name)
		} else {
			require.NoError(t, err, "Did not expect an error for scenario: %s", s.name)
			require.NotNil(t, pp, "PathPattern should not be nil for scenario: %s", s.name)
			require.Equal(t, s.expectedSegments, pp.segments, "Segments mismatch for scenario: %s", s.name)
			require.Equal(t, s.expectedNames, pp.names, "Names mismatch for scenario: %s", s.name)
			require.Equal(t, patternPath, pp.original, "Original path mismatch for scenario: %s", s.name)
		}
	}

	t.Run("Valid pattern with parameters", func(t *testing.T) {
		check(t, testScenario{
			name:       "Valid pattern with parameters",
			patternStr: "a/{p1}/b/{p2}/c",
			expectErr:  false,
			expectedSegments: []string{"", "p1", "", "p2", ""},
			expectedNames: map[string]int{"p1": 1, "p2": 3},
		})
	})

	t.Run("Valid pattern without parameters", func(t *testing.T) {
		check(t, testScenario{
			name:       "Valid pattern without parameters",
			patternStr: "a/b/c",
			expectErr:  false,
			expectedSegments: []string{"", "", ""},
			expectedNames: map[string]int{},
		})
	})

	t.Run("Pattern with empty parameter name", func(t *testing.T) {
		check(t, testScenario{
			name:       "Pattern with empty parameter name",
			patternStr: "a/{}/b",
			expectErr:  true,
			errMsg:     "path pattern has empty parameter name in segment 1",
		})
	})

	t.Run("Pattern with duplicate parameter name", func(t *testing.T) {
		check(t, testScenario{
			name:       "Pattern with duplicate parameter name",
			patternStr: "a/{p1}/b/{p1}",
			expectErr:  true,
			errMsg:     "path pattern has duplicate parameter name 'p1'",
		})
	})

	t.Run("Pattern with trailing slash", func(t *testing.T) {
		check(t, testScenario{
			name:       "Pattern with trailing slash",
			patternStr: "a/{p1}/",
			expectErr:  false,
			expectedSegments: []string{"", "p1"},
			expectedNames: map[string]int{"p1": 1},
		})
	})

	t.Run("Pattern with only parameter", func(t *testing.T) {
		check(t, testScenario{
			name:       "Pattern with only parameter",
			patternStr: "{p1}",
			expectErr:  false,
			expectedSegments: []string{"p1"},
			expectedNames: map[string]int{"p1": 0},
		})
	})

	t.Run("Complex valid pattern", func(t *testing.T) {
		check(t, testScenario{
			name:       "Complex valid pattern",
			patternStr: "/users/{id}/posts/{post_id}/comments",
			expectErr:  false,
			expectedSegments: []string{"", "id", "", "post_id", ""},
			expectedNames: map[string]int{"id": 1, "post_id": 3},
		})
	})
}

func TestUnitFormat(t *testing.T) {
	type testScenario struct {
		name         string
		patternStr   string
		params       map[string]string
		expectErr    bool
		errMsg       string
		expectedPath string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		patternPath, err := pathz.Parse(s.patternStr)
		require.NoError(t, err)

		pp, err := Parse(patternPath)
		require.NoError(t, err)

		formattedPath, err := pp.Format(s.params)

		if s.expectErr {
			require.Error(t, err, "Expected an error for scenario: %s", s.name)
			require.Contains(t, err.Error(), s.errMsg, "Error message mismatch for scenario: %s", s.name)
		} else {
			require.NoError(t, err, "Did not expect an error for scenario: %s", s.name)
			require.NotNil(t, formattedPath, "Formatted path should not be nil for scenario: %s", s.name)
			require.Equal(t, s.expectedPath, formattedPath.String(), "Formatted path mismatch for scenario: %s", s.name)
		}
	}

	t.Run("Format with all parameters provided", func(t *testing.T) {
		check(t, testScenario{
			name:         "All parameters provided",
			patternStr:   "a/{p1}/b/{p2}/c",
			params:       map[string]string{"p1": "value1", "p2": "value2"},
			expectErr:    false,
			expectedPath: "a/value1/b/value2/c",
		})
	})

	t.Run("Format with missing parameter", func(t *testing.T) {
		check(t, testScenario{
			name:       "Missing parameter",
			patternStr: "a/{p1}/b/{p2}/c",
			params:     map[string]string{"p1": "value1"},
			expectErr:  true,
			errMsg:     "parameter 'p2' not provided",
		})
	})

	t.Run("Format with empty parameter value", func(t *testing.T) {
		check(t, testScenario{
			name:       "Empty parameter value",
			patternStr: "a/{p1}/b",
			params:     map[string]string{"p1": ""},
			expectErr:  true,
			errMsg:     "parameter 'p1' cannot be empty",
		})
	})

	t.Run("Format with no parameters in pattern", func(t *testing.T) {
		check(t, testScenario{
			name:         "No parameters in pattern",
			patternStr:   "a/b/c",
			params:       map[string]string{"p1": "value1"}, // Extra params should be ignored
			expectErr:    false,
			expectedPath: "a/b/c",
		})
	})

	t.Run("Format with trailing slash", func(t *testing.T) {
		check(t, testScenario{
			name:         "Trailing slash",
			patternStr:   "a/{p1}/",
			params:       map[string]string{"p1": "value1"},
			expectErr:    false,
			expectedPath: "a/value1/",
		})
	})

	t.Run("Format with leading slash", func(t *testing.T) {
		check(t, testScenario{
			name:         "Leading slash",
			patternStr:   "/a/{p1}",
			params:       map[string]string{"p1": "value1"},
			expectErr:    false,
			expectedPath: "/a/value1",
		})
	})
}

func TestUnitMustFormat(t *testing.T) {
	t.Run("MustFormat with valid parameters", func(t *testing.T) {
		patternPath, err := pathz.Parse("a/{p1}/b")
		require.NoError(t, err)
		pp, err := Parse(patternPath)
		require.NoError(t, err)

		params := map[string]string{"p1": "test"}
		formattedPath := pp.MustFormat(params)
		require.Equal(t, "a/test/b", formattedPath.String())
	})

	t.Run("MustFormat with missing parameter panics", func(t *testing.T) {
		patternPath, err := pathz.Parse("a/{p1}/b")
		require.NoError(t, err)
		pp, err := Parse(patternPath)
		require.NoError(t, err)

		params := map[string]string{}
		require.Panics(t, func() {
			pp.MustFormat(params)
		})
	})
}

func TestUnitParsePath(t *testing.T) {
	type testScenario struct {
		name          string
		patternStr    string
		pathToParse   string
		expectErr     bool
		errMsg        string
		expectedParams map[string]string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		patternPath, err := pathz.Parse(s.patternStr)
		require.NoError(t, err)
		pp, err := Parse(patternPath)
		require.NoError(t, err)

		targetPath, err := pathz.Parse(s.pathToParse)
		require.NoError(t, err)

		params, err := pp.Parse(targetPath)

		if s.expectErr {
			require.Error(t, err, "Expected an error for scenario: %s", s.name)
			require.Contains(t, err.Error(), s.errMsg, "Error message mismatch for scenario: %s", s.name)
		} else {
			require.NoError(t, err, "Did not expect an error for scenario: %s", s.name)
			require.Equal(t, s.expectedParams, params, "Parsed parameters mismatch for scenario: %s", s.name)
		}
	}

	t.Run("Parse with matching path and parameters", func(t *testing.T) {
		check(t, testScenario{
			name:          "Matching path",
			patternStr:    "a/{p1}/b/{p2}/c",
			pathToParse:   "a/value1/b/value2/c",
			expectErr:     false,
			expectedParams: map[string]string{"p1": "value1", "p2": "value2"},
		})
	})

	t.Run("Parse with path parents mismatch", func(t *testing.T) {
		check(t, testScenario{
			name:        "Path parents mismatch",
			patternStr:  "/a/{p1}/b",
			pathToParse: "a/value1/b", // Missing leading slash
			expectErr:   true,
			errMsg:      "path parents mismatch, pattern: /a/{p1}/b, path: a/value1/b",
		})
	})

	t.Run("Parse with path length mismatch (pattern longer)", func(t *testing.T) {
		check(t, testScenario{
			name:        "Path length mismatch (pattern longer)",
			patternStr:  "a/{p1}/b/c",
			pathToParse: "a/value1/b",
			expectErr:   true,
			errMsg:      "path length mismatch, pattern: a/{p1}/b/c, path: a/value1/b",
		})
	})

	t.Run("Parse with path length mismatch (path longer)", func(t *testing.T) {
		check(t, testScenario{
			name:        "Path length mismatch (path longer)",
			patternStr:  "a/{p1}/b",
			pathToParse: "a/value1/b/c",
			expectErr:   false, // Should still parse as long as pattern matches prefix
			expectedParams: map[string]string{"p1": "value1"},
		})
	})

	t.Run("Parse with static segment mismatch", func(t *testing.T) {
		check(t, testScenario{
			name:        "Static segment mismatch",
			patternStr:  "a/{p1}/b",
			pathToParse: "a/value1/d", // 'b' vs 'd'
			expectErr:   true,
			errMsg:      "path mismatch at segment 2: expected 'b', got 'd'",
		})
	})

	t.Run("Parse with ending slash mismatch (pattern has, path not)", func(t *testing.T) {
		check(t, testScenario{
			name:        "Ending slash mismatch (pattern has)",
			patternStr:  "a/b/",
			pathToParse: "a/b",
			expectErr:   true,
			errMsg:      "ending slash mismatch, pattern: a/b/, path: a/b",
		})
	})

	t.Run("Parse with ending slash mismatch (path has, pattern not)", func(t *testing.T) {
		check(t, testScenario{
			name:        "Ending slash mismatch (path has)",
			patternStr:  "a/b",
			pathToParse: "a/b/",
			expectErr:   false, // Path can have trailing slash even if pattern doesn't
			expectedParams: map[string]string{},
		})
	})

	t.Run("Parse with no parameters", func(t *testing.T) {
		check(t, testScenario{
			name:          "No parameters",
			patternStr:    "a/b/c",
			pathToParse:   "a/b/c",
			expectErr:     false,
			expectedParams: map[string]string{},
		})
	})

	t.Run("Parse with only parameter", func(t *testing.T) {
		check(t, testScenario{
			name:          "Only parameter",
			patternStr:    "{p1}",
			pathToParse:   "value1",
			expectErr:     false,
			expectedParams: map[string]string{"p1": "value1"},
		})
	})

	t.Run("Parse with complex valid path", func(t *testing.T) {
		check(t, testScenario{
			name:          "Complex valid path",
			patternStr:    "/users/{id}/posts/{post_id}/comments",
			pathToParse:   "/users/123/posts/456/comments",
			expectErr:     false,
			expectedParams: map[string]string{"id": "123", "post_id": "456"},
		})
	})

	t.Run("Parse path longer than pattern, but still matching", func(t *testing.T) {
		check(t, testScenario{
			name:          "Path longer than pattern, but still matching",
			patternStr:    "a/{p1}",
			pathToParse:   "a/value1/b/c",
			expectErr:     false,
			expectedParams: map[string]string{"p1": "value1"},
		})
	})
}

func TestUnitMustParse(t *testing.T) {
	t.Run("MustParse with matching path", func(t *testing.T) {
		patternPath, err := pathz.Parse("a/{p1}/b")
		require.NoError(t, err)
		pp, err := Parse(patternPath)
		require.NoError(t, err)

		targetPath, err := pathz.Parse("a/test/b")
		require.NoError(t, err)

		params := pp.MustParse(targetPath)
		require.Equal(t, map[string]string{"p1": "test"}, params)
	})

	t.Run("MustParse with non-matching path panics", func(t *testing.T) {
		patternPath, err := pathz.Parse("a/{p1}/b")
		require.NoError(t, err)
		pp, err := Parse(patternPath)
		require.NoError(t, err)

		targetPath, err := pathz.Parse("a/test/d") // Mismatch 'b' vs 'd'
		require.NoError(t, err)

		require.Panics(t, func() {
			pp.MustParse(targetPath)
		})
	})
}
