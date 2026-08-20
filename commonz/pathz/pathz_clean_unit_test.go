package pathz_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/infinity6-ai/gox/commonz/pathz"
)

func TestUnitClean(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean path",
			input:    "/a/b/c",
			expected: "a/b/c",
		},
		{
			name:     "path with ..",
			input:    "/a/b/../c",
			expected: "a/c",
		},
		{
			name:     "path with .",
			input:    "/a/./b/c",
			expected: "a/b/c",
		},
		{
			name:     "path with multiple ..",
			input:    "/a/b/../../c",
			expected: "c",
		},
		{
			name:     "path with .. at the beginning",
			input:    "../a/b/c",
			expected: "a/b/c",
		},
		{
			name:     "path with trailing slash",
			input:    "/a/b/c/",
			expected: "a/b/c",
		},
		{
			name:     "path with leading and trailing slash",
			input:    "/a/b/c/",
			expected: "a/b/c",
		},
		{
			name:     "empty path",
			input:    "",
			expected: ".",
		},
		{
			name:     "root path",
			input:    "/",
			expected: ".",
		},
		{
			name:     "just ..",
			input:    "..",
			expected: ".",
		},
		{
			name:     "just .",
			input:    ".",
			expected: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := pathz.Clean(tt.input)
			require.Equal(t, tt.expected, actual)
		})
	}
}
