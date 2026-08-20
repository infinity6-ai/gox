package pathz

import "strings"

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

func Parse(path string) *Path {
	p := &Path{
		Pattern: &Pattern{
			Parts: make([]*Part, 0),
		},
		Values: make(map[string]string),
	}

	segments := splitPath(path)
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

	return p
}

func splitPath(path string) []string {
	if path == "/" {
		return []string{}
	}
	// Trim leading and trailing slashes to handle cases like "/a/b/"
	path = trimSlash(path)
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
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
