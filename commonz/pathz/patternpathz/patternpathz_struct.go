package patternpathz

import (
	"github.com/infinity6-ai/gox/commonz/pathz"
)

type PathPattern struct {
	original *pathz.Path    // a/{p1}/b/{p2}
	segments []string       // not empty on path param positions
	names    map[string]int // set of parameter names to its position
}

func Parse(pattern *pathz.Path) (*PathPattern, error) {
	panic("implement it")
}

func (p *PathPattern) String() string {
	if p == nil {
		return "<nil>"
	}
	return p.original.String()
}
