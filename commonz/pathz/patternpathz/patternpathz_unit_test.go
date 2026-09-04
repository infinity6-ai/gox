package patternpathz

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/stretchr/testify/require"
)

func TestUnitParse(t *testing.T) {
	type testScenario struct {
		name        string
		pattern     *pathz.Path
		expected    *PathPattern
		expectedErr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		pattern, err := Parse(s.pattern)

		if s.expectedErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedErr)
			return
		}

		require.NoError(t, err)
		require.Equal(t, s.expected.original, pattern.original)
		require.Equal(t, s.expected.names, pattern.names)
		require.Equal(t, s.expected.segments, pattern.segments)
	}

	scenarios := []testScenario{
		{
			name:    "valid pattern",
			pattern: pathz.MustParse("a/{p1}/b/{p2}"),
			expected: &PathPattern{
				original: pathz.MustParse("a/{p1}/b/{p2}"),
				segments: []string{"", "p1", "", "p2"},
				names:    map[string]int{"p1": 1, "p2": 3},
			},
		},
		{
			name:        "duplicate parameter name",
			pattern:     pathz.MustParse("a/{p1}/b/{p1}"),
			expectedErr: "duplicate parameter name 'p1'",
		},
		{
			name:        "empty parameter name",
			pattern:     pathz.MustParse("a/{}/b"),
			expectedErr: "empty parameter name in segment 1",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			check(t, s)
		})
	}
}

func TestUnitFormat(t *testing.T) {
	type testScenario struct {
		name        string
		pattern     *PathPattern
		params      map[string]string
		expected    *pathz.Path
		expectedErr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		path, err := s.pattern.Format(s.params)

		if s.expectedErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedErr)
			return
		}

		require.NoError(t, err)
		require.Equal(t, s.expected, path)
	}

	pattern, err := Parse(pathz.MustParse("a/{p1}/b/{p2}"))
	require.NoError(t, err)

	scenarios := []testScenario{
		{
			name:    "valid format",
			pattern: pattern,
			params:  map[string]string{"p1": "v1", "p2": "v2"},
			expected: pathz.MustParse("a/v1/b/v2"),
		},
		{
			name:        "missing parameter",
			pattern:     pattern,
			params:      map[string]string{"p1": "v1"},
			expectedErr: "parameter 'p2' not provided",
		},
		{
			name:        "empty parameter value",
			pattern:     pattern,
			params:      map[string]string{"p1": "v1", "p2": ""},
			expectedErr: "parameter 'p2' cannot be empty",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			check(t, s)
		})
	}
}

func TestUnitPathPatternParse(t *testing.T) {
	type testScenario struct {
		name        string
		pattern     *PathPattern
		path        *pathz.Path
		expected    map[string]string
		expectedErr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		params, err := s.pattern.Parse(s.path)

		if s.expectedErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedErr)
			return
		}

		require.NoError(t, err)
		require.Equal(t, s.expected, params)
	}

	pattern, err := Parse(pathz.MustParse("a/{p1}/b/{p2}"))
	require.NoError(t, err)

	scenarios := []testScenario{
		{
			name:     "matching path",
			pattern:  pattern,
			path:     pathz.MustParse("a/v1/b/v2"),
			expected: map[string]string{"p1": "v1", "p2": "v2"},
		},
		{
			name:        "length mismatch",
			pattern:     pattern,
			path:        pathz.MustParse("a/v1/b"),
			expectedErr: "path length mismatch",
		},
		{
			name:        "ending slash mismatch",
			pattern:     pattern,
			path:        pathz.MustParse("a/v1/b/v2/"),
			expectedErr: "path ending slash mismatch",
		},
		{
			name:        "literal mismatch",
			pattern:     pattern,
			path:        pathz.MustParse("c/v1/b/v2"),
			expectedErr: "path mismatch at segment 0",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			check(t, s)
		})
	}
}
