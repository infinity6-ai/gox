package pathz

import (
	"fmt"
	"path"
	"strings"
	"unicode"
)

type Path struct {
	Parts []string
}

func Parse(input string) (*Path, error) {
	if input == "" {
		return &Path{Parts: []string{}}, nil
	}

	// Validate for illegal characters
	for _, r := range input {
		if r == '\x00' {
			return nil, fmt.Errorf("path contains illegal null character")
		}
		if unicode.IsControl(r) {
			return nil, fmt.Errorf("path contains illegal control character")
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
