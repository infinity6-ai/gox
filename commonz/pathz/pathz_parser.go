package pathz

import (
	"errors"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

var ErrNavigationError = errors.New("path navigation error")

type Path struct {
	parts          []string
	parents        int
	hasEndingSlash bool
}

func New(parents int, parts []string, hasEndingSlash bool) *Path {
	var p Path
	p.set(parts, parents, hasEndingSlash)
	return &p
}

func (p *Path) Clone() *Path {
	return &Path{
		parts:          slices.Clone(p.parts),
		parents:        p.parents,
		hasEndingSlash: p.hasEndingSlash,
	}
}

func (p *Path) Parts() []string {
	return p.parts
}

func (p *Path) Parents() int {
	return p.parents
}

func (p *Path) HasEndingSlash() bool {
	return p.hasEndingSlash
}

func (p *Path) set(parts []string, parents int, hasEndingSlash bool) error {
	for i, part := range parts {
		if part == "" || strings.Trim(part, ".") == "" {
			return fmt.Errorf("path contains illegal component %d: \"%s\"", i, part)
		}
		for idx, r := range part {
			if !IsValidChar(r) {
				return fmt.Errorf("path contains illegal character %d '%c' in '%s'", idx, r, part)
			}
		}
	}
	p.parts = parts
	p.parents = parents
	p.hasEndingSlash = hasEndingSlash
	return nil
}

func (p Path) String() string {
	var sb strings.Builder
	if p.parents == -1 {
		sb.WriteString("/")
	} else {
		for i := 0; i < p.parents; i++ {
			sb.WriteString("../")
		}
	}
	sb.WriteString(strings.Join(p.parts, "/"))
	if p.hasEndingSlash && (len(p.parts) > 0 || p.parents > 0) {
		sb.WriteString("/")
	}
	return sb.String()
}

func (p *Path) Parse(input string) error {
	if input == "" {
		p.set(nil, 0, false)
		return nil
	}

	// Validate for illegal characters using IsValidChar
	for idx, r := range input {
		if !IsValidChar(r) {
			return fmt.Errorf("path contains illegal character %d '%c' in '%s'", idx, r, input)
		}
	}

	isAbsolute := strings.HasPrefix(input, "/")
	hasEndingSlash := strings.HasSuffix(input, "/") && len(input) > 1

	// Clean the path using the standard library
	cleanedPath := path.Clean(input)

	if cleanedPath == "." {
		p.set(nil, 0, hasEndingSlash)
		return nil
	}

	// path.Clean("/a/../b") -> "/b"
	// path.Clean("././.") -> "."
	// path.Clean("/") -> "/"

	// if cleanedPath == "." {
	// return &ret, nil
	// }]
	var parts []string
	parents := 0
	if cleanedPath == "/" {
		parents = -1
	} else {
		parts = strings.Split(cleanedPath, "/")
		if isAbsolute {
			if parts[0] != "" {
				panic(fmt.Errorf("it was not supposed to happen: %s", cleanedPath))
			}
			parents = -1
			parts = parts[1:]
		} else {
			parents, parts = countParents(parts)
		}
	}

	err := p.set(parts, parents, hasEndingSlash)
	if err != nil {
		return err
	}
	return nil
}

func MustParse(input string) *Path {
	ret, err := Parse(input)
	errorz.Check(err)
	return ret
}

func Parse(input string) (*Path, error) {
	var ret Path
	err := ret.Parse(input)
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
