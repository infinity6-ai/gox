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

	oldLen := len(opts.Out)

	var limitedReader io.Reader = r
	if opts.Max > 0 {
		limitedReader = io.LimitReader(r, int64(opts.Max))
	}

	totalRead := 0
	for {
		// check context in loop
		select {
		case <-ctx.Done():
			opts.Out = opts.Out[:oldLen+totalRead] // reslice to what we have read
			return ctx.Err()
		default:
		}

		buf := opts.grow(defaultBufferCap)
		n, err := limitedReader.Read(buf)

		totalRead += n
		opts.Out = opts.Out[:oldLen+totalRead] // Reslice to actual data read.

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	if totalRead < opts.Min {
		return io.ErrUnexpectedEOF
	}

	return nil
}
