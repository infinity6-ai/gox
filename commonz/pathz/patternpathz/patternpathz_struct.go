package patternpathz

import (
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/pathz"
)

type PathPattern struct {
	original *pathz.Path    // a/{p1}/b/{p2}
	segments []string       // not empty on path param positions
	names    map[string]int // set of parameter names to its position
}

func Parse(pattern *pathz.Path) (*PathPattern, error) {
	parts := pattern.Parts()
	p := &PathPattern{
		original: pattern,
		segments: make([]string, len(parts)),
		names:    make(map[string]int),
	}

	for i, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			name := strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}")
			if name == "" {
				return nil, fmt.Errorf("path pattern has empty parameter name in segment %d", i)
			}
			if _, ok := p.names[name]; ok {
				return nil, fmt.Errorf("path pattern has duplicate parameter name '%s'", name)
			}
			p.names[name] = i
			p.segments[i] = name
		}
	}

	return p, nil
}

func (p *PathPattern) String() string {
	if p == nil {
		return "<nil>"
	}
	return p.original.String()
}
