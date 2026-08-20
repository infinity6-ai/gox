package pathz

import (
	"fmt"
	"strings"
	"unicode"
)

type Part struct {
	Name        string
	Placeholder bool
}

type Pattern struct {
	Parts []*Part
}

type Path struct {
	Pattern *Pattern
	Values  map[string]string
}

func (p *Path) String() string {
	var sb strings.Builder
	for _, part := range p.Pattern.Parts {
		sb.WriteString("/")
		if part.Placeholder {
			if val, ok := p.Values[part.Name]; ok {
				sb.WriteString(val)
			} else {
				sb.WriteString("{")
				sb.WriteString(part.Name)
				sb.WriteString("}")
			}
		} else {
			sb.WriteString(part.Name)
		}
	}
	return sb.String()
}

func Parse(path string) (*Path, error) {
	p := &Path{
		Pattern: &Pattern{
			Parts: make([]*Part, 0),
		},
		Values: make(map[string]string),
	}

	segments, err := splitPath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to split path: %w", err)
	}
	for _, segment := range segments {
		if segment == "" {
			continue
		}
		if len(segment) > 2 && segment[0] == '{' && segment[len(segment)-1] == '}' {
			p.Pattern.Parts = append(p.Pattern.Parts, &Part{
				Name:        segment[1 : len(segment)-1],
				Placeholder: true,
			})
		} else {
			p.Pattern.Parts = append(p.Pattern.Parts, &Part{
				Name:        segment,
				Placeholder: false,
			})
		}
	}

	return p, nil
}

func splitPath(path string) ([]string, error) {
	path = Clean(path)
	if path == "." {
		return []string{}, nil
	}
	segments := strings.Split(path, "/")
	for _, segment := range segments {
		if len(segment) > 2 && segment[0] == '{' && segment[len(segment)-1] == '}' {
			continue
		}
		if !isSafe(segment) {
			return nil, fmt.Errorf("illegal character in path segment: %s", segment)
		}
	}
	return segments, nil
}

func trimSlash(s string) string {
	if strings.HasPrefix(s, "/") {
		s = s[1:]
	}
	if strings.HasSuffix(s, "/") {
		s = s[:len(s)-1]
	}
	return s
}

func isSafe(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '-' && r != '_' && r != '.' {
			return false
		}
	}
	return true
}
