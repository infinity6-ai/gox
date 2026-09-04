package patternpathz

import "github.com/infinity6-ai/gox/commonz/pathz"

type PathPattern struct {
	originalPattern *pathz.Path
	segments        []string       // not empty on path param positions
	paramNames      map[string]int // set of parameter names to its position
}

func Parse(pattern *pathz.Path) (*PathPattern, error) {
	panic("implement it")
}

func (p *PathPattern) Format(params map[string]string) (*pathz.Path, error) {
	panic("implement it")
}

func (p *PathPattern) MustFormat(params map[string]string) *pathz.Path {
	panic("implement it")
}

func (p *PathPattern) Match(path *pathz.Path) bool {
	panic("implement it")
}

func (p *PathPattern) MustMatch(path *pathz.Path) bool {
	panic("implement it")
}

func (p *PathPattern) String() string {
	if p == nil {
		return "<nil>"
	}
	return p.originalPattern.String()
}
