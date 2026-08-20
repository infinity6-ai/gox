package pathz_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/infinity6-ai/gox/commonz/pathz"
)

func TestUnitParseBasic(t *testing.T) {
	tests := []struct {
		name                   string
		input                  string
		expectedParts          []string
		expectedParents        int
		expectedHasEndingSlash bool
		expectedError          string
	}{
		{
			name:            "absolute path",
			input:           "/a/b/c",
			expectedParts:   []string{"a", "b", "c"},
			expectedParents: -1,
		},
		{
			name:            "relative path",
			input:           "a/b/c",
			expectedParts:   []string{"a", "b", "c"},
			expectedParents: 0,
		},
		{
			name:            "path with current dir",
			input:           "a/./b/c",
			expectedParts:   []string{"a", "b", "c"},
			expectedParents: 0,
		},
		{
			name:            "path with parent dir",
			input:           "a/b/../c",
			expectedParts:   []string{"a", "c"},
			expectedParents: 0,
		},
		{
			name:            "path starting with parent dir (relative)",
			input:           "../a/b",
			expectedParts:   []string{"a", "b"},
			expectedParents: 1,
		},
		{
			name:            "path starting with parent dir (absolute)",
			input:           "/../a/b",
			expectedParts:   []string{"a", "b"},
			expectedParents: -1,
		},
		{
			name:            "path with multiple parent dir navigations",
			input:           "a/b/c/../../d",
			expectedParts:   []string{"a", "d"},
			expectedParents: 0,
		},
		{
			name:            "path with excessive parent dir navigations",
			input:           "a/../../b",
			expectedParts:   []string{"b"},
			expectedParents: 1,
		},
		{
			name:            "empty path",
			input:           "",
			expectedParts:   []string{},
			expectedParents: 0,
		},
		{
			name:                   "path with only slashes",
			input:                  "///",
			expectedParts:          []string{},
			expectedParents:        -1,
			expectedHasEndingSlash: true,
		},
		{
			name:            "path with only current dir",
			input:           "././.",
			expectedParts:   []string{},
			expectedParents: 0,
		},
		{
			name:                   "path with only parent dir",
			input:                  "../../",
			expectedParts:          []string{},
			expectedParents:        2,
			expectedHasEndingSlash: true,
		},
		{
			name:          "path with illegal characters - null byte",
			input:         "a\x00b/c",
			expectedError: "path contains illegal character: '\x00'",
		},
		{
			name:          "path with illegal characters - control character",
			input:         "a\tb/c",
			expectedError: "path contains illegal character: '\t'",
		},
		{
			name:                   "path with trailing slash",
			input:                  "a/b/c/",
			expectedParts:          []string{"a", "b", "c"},
			expectedParents:        0,
			expectedHasEndingSlash: true,
		},
		{
			name:            "path with leading slash",
			input:           "/a/b/c",
			expectedParts:   []string{"a", "b", "c"},
			expectedParents: -1,
		},
		{
			name:            "path with mixed slashes",
			input:           "//a/b///c",
			expectedParts:   []string{"a", "b", "c"},
			expectedParents: -1,
		},
		{
			name:            "complex path",
			input:           "/./a/../b/./c/../../d/e/.",
			expectedParts:   []string{"d", "e"},
			expectedParents: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := pathz.Parse(tt.input)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				require.Nil(t, p)
			} else {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.Equal(t, tt.expectedParts, p.Parts)
				require.Equal(t, tt.expectedParents, p.Parents)
				require.Equal(t, tt.expectedHasEndingSlash, p.HasEndingSlash)
			}
		})
	}
}
