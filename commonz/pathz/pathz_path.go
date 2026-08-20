package pathz

import (
	"errors"
	"fmt"
	"path"
	"strings"
	// "unicode" // No longer needed as IsValidChar handles this
)

type Path struct {
	Parts          []string
	Parents        int
	HasEndingSlash bool
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

	isAbsolute := strings.HasPrefix(input, "/")
	hasEndingSlash := strings.HasSuffix(input, "/") && len(input) > 1

	// Clean the path using the standard library
	cleanedPath := path.Clean(input)

	// path.Clean("/a/../b") -> "/b"
	// path.Clean("././.") -> "."
	// path.Clean("/") -> "/"
	// path.Clean("") -> "."
	// path.Clean("../../a") -> "../../a"

	parents := 0
	var parts []string

	if isAbsolute {
		parents = -1
		if cleanedPath == "/" {
			parts = []string{}
		} else {
			parts = strings.Split(strings.TrimPrefix(cleanedPath, "/"), "/")
		}
	} else { // relative
		pathSegments := strings.Split(cleanedPath, "/")

		i := 0
		for i < len(pathSegments) && pathSegments[i] == ".." {
			parents++
			i++
		}

		remainingSegments := pathSegments[i:]
		if len(remainingSegments) == 1 && remainingSegments[0] == "." {
			parts = []string{}
		} else {
			parts = remainingSegments
		}
	}

	if len(parts) > 0 {
		for _, part := range parts {
			if part == "" {
				panic(errors.New("Should not happen with path.Clean, but as a safeguard."))
			}
			allDots := true
			for _, r := range part {
				if r != '.' {
					allDots = false
					break
				}
			}
			if allDots {
				return nil, fmt.Errorf("path contains illegal component: \"%s\"", part)
			}
		}
	}

	return &Path{
		Parts:          parts,
		Parents:        parents,
		HasEndingSlash: hasEndingSlash,
	}, nil
}
