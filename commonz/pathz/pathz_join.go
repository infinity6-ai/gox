package pathz

import "errors"

var ErrEscaped = errors.New("path escaoed error")

func (p *Path) Join(others ...*Path) (*Path, error) {
	panic("implement it")
}
