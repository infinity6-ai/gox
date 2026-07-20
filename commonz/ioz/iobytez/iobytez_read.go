package iobytez

import (
	"context"
	"io"
	"slices"
)

const (
	defaultBufferCap = 4096
)

type Options struct {
	Min int
	Max int
	Out []byte
}

func (o *Options) grow(size int) []byte {
	oldLen := len(o.Out)
	o.Out = slices.Grow(o.Out, size)
	o.Out = o.Out[:oldLen+size]
	return o.Out[oldLen:]
}

func Read(ctx context.Context, r io.Reader, opts *Options) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c := defaultBufferCap
	if c < opts.Min {
		c = opts.Min
	}
	if opts.Max > 0 && c > opts.Max {
		c = opts.Max
	}
	opts.grow(c)

}
