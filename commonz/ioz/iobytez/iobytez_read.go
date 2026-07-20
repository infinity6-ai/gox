package iobytez

import (
	"context"
	"io"
	"slices"
)

const (
	defaultGrowSize = 4096 // Renamed from defaultBufferCap
)

type Options struct {
	Min int
	Max int
	Out []byte
	GrowSize int // New field
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

	var growSize = opts.GrowSize // Use opts.GrowSize
	if growSize == 0 {           // If not set, use default
		growSize = defaultGrowSize
	}

	var limitedReader io.Reader = r
	if opts.Max > 0 {
		limitedReader = io.LimitReader(r, int64(opts.Max))
	}

	totalRead := 0
	var eofEncountered bool
	for {
		// check context in loop
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		lenBeforeGrow := len(opts.Out)

		// Calculate remaining capacity if Max is set
		sizeToGrow := growSize
		if opts.Max > 0 {
			remaining := opts.Max - len(opts.Out)
			if remaining < sizeToGrow {
				sizeToGrow = remaining
			}
		}

		buf := opts.grow(sizeToGrow) // Use calculated sizeToGrow
		n, err := limitedReader.Read(buf)

		// Reslice to the actual number of bytes read in this iteration.
		opts.Out = opts.Out[:lenBeforeGrow+n]
		totalRead += n

		if err == io.EOF {
			eofEncountered = true
			break
		}
		if err != nil {
			return err
		}
	}

	if totalRead < opts.Min {
		return io.ErrUnexpectedEOF
	}

	if eofEncountered && totalRead == 0 {
		return io.EOF
	}

	return nil
}