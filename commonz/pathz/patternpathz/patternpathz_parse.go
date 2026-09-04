package patternpathz

import (
	"fmt"
	"slices"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/pathz"
)

func (p *PathPattern) Parse(pt *pathz.Path) (map[string]string, *pathz.Path, error) {
	if p.original.Parents() != pt.Parents() {
		return nil, nil, fmt.Errorf("path parents mismatch, pattern: %s, path: %s", p, pt)
	}

	pParts := p.original.Parts()
	ptParts := pt.Parts()

	if len(pParts) > len(ptParts) {
		return nil, nil, fmt.Errorf("path length mismatch, pattern: %s, path: %s", p, pt)
	}

	if len(pParts) == len(ptParts) && p.original.HasEndingSlash() && !pt.HasEndingSlash() {
		return nil, nil, fmt.Errorf("ending slash mismatch, pattern: %s, path: %s", p, pt)
	}

	params := make(map[string]string)
	for i, name := range p.segments {
		if name != "" {
			params[name] = ptParts[i]
		} else if pParts[i] != ptParts[i] {
			return nil, nil, fmt.Errorf("path mismatch at segment %d: expected '%s', got '%s'", i, pParts[i], ptParts[i])
		}
	}

	suffixParts := slices.Clone(ptParts[len(p.segments):])
	if len(suffixParts) > 0 {
		return params, pathz.New(pt.Parents(), suffixParts, pt.HasEndingSlash()), nil
	}

	return params, nil, nil
}

func (p *PathPattern) MustParse(pt *pathz.Path) map[string]string {
	params, err := p.Parse(pt)
	errorz.Check(err)
	return params
}
