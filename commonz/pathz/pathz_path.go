package pathz

import (
	"fmt"
	"path"
	"strings"
	// "unicode" // No longer needed as IsValidChar handles this
)

type Path struct {
	Parts []string
}

func Parse(input string) (*Path, error) {
	if input == "" {
		return &Path{Parts: []string{}}, nil
	}

	// Validate for illegal characters using IsValidChar
	for _, r := range input {
		if !IsValidChar(r) {
			return nil, fmt.Errorf("path contains illegal character: '%c'", r)
		}
	}

	// Clean the path using the standard library
	cleanedPath := path.Clean(input)

	// path.Clean("/a/../b") -> "/b"
	// path.Clean("././.") -> "."
	// path.Clean("/") -> "/"
	// path.Clean("") -> "."

	// If the cleaned path is "." or "/", it should result in an empty set of parts
	if cleanedPath == "." || cleanedPath == "/" {
		return &Path{Parts: []string{}}, nil
	}

	// Remove leading and trailing slashes for splitting
	cleanedPath = strings.Trim(cleanedPath, "/")

	parts := strings.Split(cleanedPath, "/")
	return &Path{Parts: parts}, nil
}
