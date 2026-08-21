package pathz

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

type Path struct {
	Parts          string
	Parents        int
	HasEndingSlash bool
}

func (p *Path) String() string {
	if p == nil {
		return ""
	}
	var sb strings.Builder
	if p.Parents == -1 {
		sb.WriteString("/")
	} else {
		for i := 0; i < p.Parents; i++ {
			sb.WriteString("../")
		}
	}
	sb.WriteString(p.Parts)
	if p.HasEndingSlash && len(p.Parts) > 0 {
		sb.WriteString("/")
	}
	return sb.String()
}

func Parse(input string) (*Path, error) {
	if input == "" {
		return &Path{Parts: ""}, nil
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

	if cleanedPath == "." {
		return &Path{Parts: ""}, nil
	}
	if cleanedPath == "/" {
		return &Path{Parts: cleanedPath[1:], Parents: -1}, nil
	}

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
		Parts:          strings.Join(parts, "/"),
		Parents:        parents,
		HasEndingSlash: hasEndingSlash,
	}, nil
}

func IsValidChar(r rune) bool {
	if r >= 'a' && r <= 'z' {
		return true
	}
	if r >= 'A' && r <= 'Z' {
		return true
	}
	if r >= '0' && r <= '9' {
		return true
	}
	if r == '-' || r == '_' || r == '.' || r == '/' || r == '=' || r == '{' || r == '}' {
		return true
	}
	return false
}
