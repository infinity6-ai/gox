package iobytez

import (
	"context"
	"io"
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
	o.Out = append(o.Out, make([]byte, size)...)
	return o.Out[len(o.Out)-size : len(o.Out)]
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
