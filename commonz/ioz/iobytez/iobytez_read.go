package iobytez

import (
	"context"
	"io"
	"time"
)

type Options struct {
	Out     []byte
	Min     int
	Max     int
	Timeout time.Duration

	Len int
}

func Read(ctx context.Context, r io.Reader, opts *Options) error {
	if opts.Out == nil {
		opts.Out = make([]byte, 0, 4096)
	}

	// timeout 0 means no timeout
	// return io.EOF should always return Len 0
	// return less than min must return io.ErrUnexpectedEOF

	panic("implement it")
}
