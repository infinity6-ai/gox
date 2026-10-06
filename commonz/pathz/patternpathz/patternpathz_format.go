package patternpathz

import (
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/pathz"
)

func (p *Pattern) Format(params map[string]string) (*pathz.Path, error) {
	originalParts := p.original.Parts()
	var newParts []string
	hasEndingSlash := p.original.HasEndingSlash()

	for i, name := range p.segments {
		if i == p.catchAll {
			value, ok := params[name]
			if !ok {
				return nil, fmt.Errorf("parameter '%s' not provided", name)
			}
			if value != "" {
				if strings.HasSuffix(value, "/") {
					hasEndingSlash = true
				}
				trimmed := strings.Trim(value, "/")
				if trimmed != "" {
					newParts = append(newParts, strings.Split(trimmed, "/")...)
				}
			}
		} else if name != "" {
			value, ok := params[name]
			if !ok {
				return nil, fmt.Errorf("parameter '%s' not provided", name)
			}
			if value == "" {
				return nil, fmt.Errorf("parameter '%s' cannot be empty", name)
			}
			newParts = append(newParts, value)
		} else {
			newParts = append(newParts, originalParts[i])
		}
	}

	return pathz.New(p.original.Parents(), newParts, hasEndingSlash), nil
}

func (p *Pattern) MustFormat(params map[string]string) *pathz.Path {
	path, err := p.Format(params)
	errorz.Check(err)
	return path
}
