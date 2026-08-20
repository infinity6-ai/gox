package pathz

import (
	"fmt"
	"strings"
	"unicode"
)

type Path struct {
	Parts []string
}

func Parse(path string) (*Path, error) {
	if path == "" {
		return &Path{Parts: []string{}}, nil
	}

	// Validate for illegal characters
	for _, r := range path {
		if r == '\x00' {
			return nil, fmt.Errorf("path contains illegal null character")
		}
		if unicode.IsControl(r) {
			return nil, fmt.Errorf("path contains illegal control character")
		}
	}

	// Split the path into segments and clean up
	segments := strings.Split(path, "/")
	cleanedSegments := make([]string, 0) // Initialize as an empty slice

	isAbsolutePath := strings.HasPrefix(path, "/")

	for _, segment := range segments {
		if segment == "" || segment == "." {
			continue
		}
		if segment == ".." {
			if len(cleanedSegments) > 0 && cleanedSegments[len(cleanedSegments)-1] != ".." {
				// Pop the last segment if it's not ".."
				cleanedSegments = cleanedSegments[:len(cleanedSegments)-1]
			} else if !isAbsolutePath { // Only add ".." for relative paths that go above the initial path
				cleanedSegments = append(cleanedSegments, "..")
			}
			continue
		}
		cleanedSegments = append(cleanedSegments, segment)
	}

	return &Path{Parts: cleanedSegments}, nil
}
