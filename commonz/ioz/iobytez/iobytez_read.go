package iobytez

import (
	"bytes"
	"context"
	"io"
)

const (
	defaultBufferCap = 4096
)

type Options struct {
	Min int
	Max int
	Out *bytes.Buffer
}

func (o *Options) fix() {
	if o.Out == nil {
		c := defaultBufferCap
		if c < o.Min {
			c = o.Min
		}
		if o.Max > 0 && c > o.Max {
			c = o.Max
		}
		o.Out = bytes.NewBuffer(make([]byte, c))
	}
}

func Read(ctx context.Context, r io.Reader, opts *Options) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	opts.fix()
}
