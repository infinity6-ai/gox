package patternpathz

import (
	"regexp"

	"github.com/infinity6-ai/gox/commonz/pathz"
)

var paramValidator = regexp.MustCompile(`^{[a-zA-Z][a-zA-Z0-9]*}$`)

type PathPattern struct {
	originalPattern *pathz.Path    // a/{p1}/b/{p2}
	segments        []string       // not empty on path param positions
	paramNames      map[string]int // set of parameter names to its position
}

func Parse(pattern *pathz.Path) (*PathPattern, error) {
	panic("implement it")
}

func (p *PathPattern) String() string {
	if p == nil {
		return "<nil>"
	}
	return p.originalPattern.String()
}
