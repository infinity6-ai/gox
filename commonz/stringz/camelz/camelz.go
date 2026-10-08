package camelz

import (
	"errors"
	"fmt"
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
	panic("implement")
}
