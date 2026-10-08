package camelz

import (
	"errors"
	"fmt"
	"strings"
)

var ErrUnsupported = errors.New("unsupported")

type Parsed struct {
	parts []string
}

// Parse cammel, snake lower or kebab lower into lower case parts
// Conercase: ABC = []string{"b", "b", "c"}
func P(s string) (*Parsed, error) {
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
	return strings.Join(p.parts, "-")
}
