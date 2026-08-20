package pathz_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/infinity6-ai/gox/commonz/pathz"
)

func TestUnitParseBasic(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedParts []string
		expectedError string
	}{
		{
			name:          "absolute path",
			input:         "/a/b/c",
			expectedParts: []string{"a", "b", "c"},
		},
		{
			name:          "relative path",
			input:         "a/b/c",
			expectedParts: []string{"a", "b", "c"},
		},
		{
			name:          "path with current dir",
			input:         "a/./b/c",
			expectedParts: []string{"a", "b", "c"},
		},
		{
			name:          "path with parent dir",
			input:         "a/b/../c",
			expectedParts: []string{"a", "c"},
		},
		{
			name:          "path starting with parent dir (relative)",
			input:         "../a/b",
			expectedParts: []string{"..", "a", "b"},
		},
		{
			name:          "path starting with parent dir (absolute)",
			input:         "/../a/b",
			expectedParts: []string{"a", "b"},
		},
		{
			name:          "path with multiple parent dir navigations",
			input:         "a/b/c/../../d",
			expectedParts: []string{"a", "d"},
		},
		{
			name:          "path with excessive parent dir navigations",
			input:         "a/../../b",
			expectedParts: []string{"..", "b"},
		},
		{
			name:          "empty path",
			input:         "",
			expectedParts: []string{},
		},
		{
			name:          "path with only slashes",
			input:         "///",
			expectedParts: []string{},
		},
		{
			name:          "path with only current dir",
			input:         "././.",
			expectedParts: []string{},
		},
		{
			name:          "path with only parent dir",
			input:         "../../",
			expectedParts: []string{"..", ".."},
		},
		{
			name:          "path with illegal characters - null byte",
			input:         "a\x00b/c",
			expectedError: "path contains illegal null character",
		},
		{
			name:          "path with illegal characters - control character",
			input:         "a\tb/c",
			expectedError: "path contains illegal control character",
		},
		{
			name:          "path with trailing slash",
			input:         "a/b/c/",
			expectedParts: []string{"a", "b", "c"},
		},
		{
			name:          "path with leading slash",
			input:         "/a/b/c",
			expectedParts: []string{"a", "b", "c"},
		},
		{
			name:          "path with mixed slashes",
			input:         "//a/b///c",
			expectedParts: []string{"a", "b", "c"},
		},
		{
			name:          "complex path",
			input:         "/./a/../b/./c/../../d/e/.",
			expectedParts: []string{"d", "e"},
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
			}
		})
	}
}
