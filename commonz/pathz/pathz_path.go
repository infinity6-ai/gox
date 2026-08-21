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

	// if cleanedPath == "/" {
	// 	return &Path{Parts: cleanedPath[1:], Parents: -1, HasEndingSlash: hasEndingSlash}, nil
	// }

	// parents := 0
	// var parts []string

	// if isAbsolute {
	// 	parents = -1
	// 	if cleanedPath == "/" {
	// 		parts = []string{}
	// 	} else {
	// 		parts = strings.Split(strings.TrimPrefix(cleanedPath, "/"), "/")
	// 	}
	// } else { // relative
	// 	pathSegments := strings.Split(cleanedPath, "/")

	// 	i := 0
	// 	for i < len(pathSegments) && pathSegments[i] == ".." {
	// 		parents++
	// 		i++
	// 	}

	// 	remainingSegments := pathSegments[i:]
	// 	if len(remainingSegments) == 1 && remainingSegments[0] == "." {
	// 		parts = []string{}
	// 	} else {
	// 		parts = remainingSegments
	// 	}
	// }

	// if len(parts) > 0 {
	// 	for _, part := range parts {
	// 		if part == "" {
	// 			panic(errors.New("Should not happen with path.Clean, but as a safeguard."))
	// 		}
	// 		allDots := true
	// 		for _, r := range part {
	// 			if r != '.' {
	// 				allDots = false
	// 				break
	// 			}
	// 		}
	// 		if allDots {
	// 			return nil, fmt.Errorf("path contains illegal component: \"%s\"", part)
	// 		}
	// 	}
	// }

	// return &Path{
	// 	Parts:          strings.Join(parts, "/"),
	// 	Parents:        parents,
	// 	HasEndingSlash: hasEndingSlash,
	// }, nil
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
