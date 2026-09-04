package patternpathz

import (
	"fmt"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/pathz"
)

func (p *Pattern) Format(params map[string]string) (*pathz.Path, error) {
	originalParts := p.original.Parts()
	newParts := make([]string, len(originalParts))

	for i, name := range p.segments {
		if name != "" {
			value, ok := params[name]
			if !ok {
				return nil, fmt.Errorf("parameter '%s' not provided", name)
			}
			if value == "" {
				return nil, fmt.Errorf("parameter '%s' cannot be empty", name)
			}
			newParts[i] = value
		} else {
			newParts[i] = originalParts[i]
		}
	}

	return pathz.New(p.original.Parents(), newParts, p.original.HasEndingSlash()), nil
}

func (p *Pattern) MustFormat(params map[string]string) *pathz.Path {
	path, err := p.Format(params)
	errorz.Check(err)
	return path
}
