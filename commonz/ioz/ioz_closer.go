package ioz

import (
	"fmt"
	"io"
)

type closer struct {
	f func() error
}

func (c *closer) Close() error {
	if c == nil || c.f == nil {
		return nil
	}
	if err := c.f(); err != nil {
		return fmt.Errorf("failed to close: %w", err)
	}
	return nil
}

func Closer(f func() error) io.Closer {
	return &closer{f: f}
}

func CloserV(f func()) io.Closer {
	return &closer{f: func() error {
		if f != nil {
			f()
		}
		return nil
	}}
}
