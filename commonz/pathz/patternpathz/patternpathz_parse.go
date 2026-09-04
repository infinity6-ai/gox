package patternpathz

import (
	"fmt"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/pathz"
)

func (p *PathPattern) Parse(pt *pathz.Path) (map[string]string, error) {
	if p.original.Parents() != pt.Parents() {
		return nil, fmt.Errorf("path parents mismatch")
	}

	pParts := p.original.Parts()
	ptParts := pt.Parts()

	if len(pParts) != len(ptParts) {
		return nil, fmt.Errorf("path length mismatch")
	}

	if p.original.HasEndingSlash() != pt.HasEndingSlash() {
		return nil, fmt.Errorf("path ending slash mismatch")
	}

	params := make(map[string]string)
	for i, name := range p.segments {
		if name != "" {
			params[name] = ptParts[i]
		} else {
			if pParts[i] != ptParts[i] {
				return nil, fmt.Errorf("path mismatch at segment %d: expected '%s', got '%s'", i, pParts[i], ptParts[i])
			}
		}
	}

	return params, nil
}

func (p *PathPattern) MustParse(pt *pathz.Path) map[string]string {
	params, err := p.Parse(pt)
	errorz.Check(err)
	return params
}
