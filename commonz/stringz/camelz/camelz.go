package camelz

import (
	"errors"
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

var ErrUnsupported = errors.New("unsupported")

type Parsed struct {
	parts []string
}

func P(s string) *Parsed {
	ret, err := Parse(s)
	errorz.Check(err)
	return ret
}

// Parse cammel, snake lower or kebab lower into lower case parts
// Conercase: ABC = []string{"b", "b", "c"}
func Parse(s string) (*Parsed, error) {
	return nil, fmt.Errorf("%w: %s", ErrUnsupported, s)
}

func (p *Parsed) String() string {
	return p.QL()
}

// To cammel string
func (p *Parsed) C() string {
	panic("implement")
}

// To snake upper string
func (p *Parsed) SU() string {
	panic("implement")
}

// To snake lower string
func (p *Parsed) SL() string {
	panic("implement")
}

// To kebab upper string
func (p *Parsed) QU() string {
	panic("implement")
}

// To kebab lower string
func (p *Parsed) QL() string {
	if p == nil {
		return ""
	}
	return strings.Join(p.parts, "-")
}
