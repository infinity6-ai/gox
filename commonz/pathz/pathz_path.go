package pathz

import (
	"fmt"
	"path"
	"strings"
	// "unicode" // No longer needed as IsValidChar handles this
)

type Path struct {
	Parts          []string
	Absolute       bool
	HasEndingSlash bool
}

func Parse(input string) (*Path, error) {
	if input == "" {
		return &Path{}, nil
	}

	// Validate for illegal characters using IsValidChar
	for _, r := range input {
		if !IsValidChar(r) {
			return nil, fmt.Errorf("path contains illegal character: '%c'", r)
		}
	}

	isAbsolute := strings.HasPrefix(input, "/")
	hasEndingSlash := strings.HasSuffix(input, "/") && len(input) > 1

	// Clean the path using the standard library
	cleanedPath := path.Clean(input)

	// path.Clean("/a/../b") -> "/b"
	// path.Clean("././.") -> "."
	// path.Clean("/") -> "/"
	// path.Clean("") -> "."

	var parts []string
	// If the cleaned path is "." or "/", it should result in an empty set of parts
	if cleanedPath != "." && cleanedPath != "/" {
		// Remove leading and trailing slashes for splitting
		cleanedPath = strings.Trim(cleanedPath, "/")
		parts = strings.Split(cleanedPath, "/")
	}

	// Validate for parts with "..." or more dots
	for _, part := range parts {
		if len(part) >= 3 && strings.Trim(part, ".") == "" {
			return nil, fmt.Errorf("path contains illegal component: %q", part)
		}
	}

	return &Path{
		Parts:          parts,
		Absolute:       isAbsolute,
		HasEndingSlash: hasEndingSlash,
	}, nil
}
