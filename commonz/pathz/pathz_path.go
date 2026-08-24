package pathz

import (
	"fmt"
	"path"
	"strings"
)

type Path struct {
	Parts          string
	Parents        int
	HasEndingSlash bool
}

func (p *Path) SetParts(parts []string) error {
	n := make([]string, len(parts))
	for i, part := range parts {
		if len(part) > 2 && strings.Trim(part, ".") == "" {
			return fmt.Errorf("path contains illegal component: \"%s\"", part)
		}
		for _, r := range part {
			if !IsValidChar(r) {
				return fmt.Errorf("path contains illegal character: '%c'", r)
			}
		}
		n[i] = part
	}
	p.Parts = strings.Join(n, "/")
	return nil
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
	for idx, r := range input {
		if !IsValidChar(r) {
			return nil, fmt.Errorf("path contains illegal character %d '%c' in '%s'", idx, r, input)
		}
	}

	var ret Path

	isAbsolute := strings.HasPrefix(input, "/")
	hasEndingSlash := strings.HasSuffix(input, "/") && len(input) > 1
	ret.HasEndingSlash = hasEndingSlash

	// Clean the path using the standard library
	cleanedPath := path.Clean(input)

	// path.Clean("/a/../b") -> "/b"
	// path.Clean("././.") -> "."
	// path.Clean("/") -> "/"
	// path.Clean("") -> "."
	// path.Clean("../../a") -> "../../a"

	// if cleanedPath == "." {
	// return &ret, nil
	// }
	parts := strings.Split(cleanedPath, "/")
	if isAbsolute {
		if parts[0] != "" {
			panic(fmt.Errorf("it was not supposed to happen: %s", cleanedPath))
		}
		ret.Parents = -1
		parts = parts[1:]
	} else {
		ret.Parents, parts = countParents(parts)
	}
	err := ret.SetParts(parts)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func countParents(parts []string) (int, []string) {
	parents := 0
	count := 0
	for _, part := range parts {
		if part == "." {
			count++
			continue
		}
		if part == ".." {
			count++
			parents++
			continue
		}
		break
	}
	return parents, parts[count:]
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
