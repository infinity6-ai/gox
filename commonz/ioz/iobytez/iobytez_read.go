package iobytez

import (
	"context"
	"fmt"
	"io"
	"slices"
)

const (
	DefaultGrowSize = 4096
)

// Options defines the parameters for the Read function, allowing for fine-grained
// control over the read operation.
type Options struct {
	// Min is the minimum number of bytes to read. If the reader returns an EOF
	// before Min bytes are read, Read returns io.ErrUnexpectedEOF.
	Min int
	// Max is the maximum number of bytes to read. Read will not read more than
	// Max bytes from the reader.
	Max int
	// Out is the buffer where the read data is stored. If Out has a non-zero
	// capacity, Read will append to it.
	Out []byte
	// GrowSize specifies the size by which the output buffer should grow when
	// more capacity is needed. If zero, a default size is used.
	GrowSize int
}

func (o *Options) grow(size int) []byte {
	oldLen := len(o.Out)
	o.Out = slices.Grow(o.Out, size)
	o.Out = o.Out[:oldLen+size]
	return o.Out[oldLen:]
}

// Read reads from r into opts.Out until the reader returns io.EOF or an error.
// It provides advanced control over the read operation through the Options struct.
//
// The function reads from r and appends the data to opts.Out. It respects the
// Min and Max fields of the opts parameter to control the amount of data read.
// The read operation can be cancelled via the context.
//
// Read returns the total number of bytes read and an error if one occurred.
// If the number of bytes read is less than opts.Min, it returns io.ErrUnexpectedEOF.
// If the reader is empty and no bytes are read, it returns io.EOF.
func Read(ctx context.Context, r io.Reader, opts *Options) (int, error) {
	select {
	case <-ctx.Done():
		return 0, fmt.Errorf("read canceled: %w", ctx.Err())
	default:
	}

	var growSize = opts.GrowSize // Use opts.GrowSize
	if growSize == 0 {           // If not set, use default
		growSize = DefaultGrowSize
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
			return totalRead, fmt.Errorf("read canceled: %w", ctx.Err())
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
			return totalRead, fmt.Errorf("failed to read: %w", err)
		}
	}

	if totalRead < opts.Min {
		return totalRead, io.ErrUnexpectedEOF
	}

	if eofEncountered && totalRead == 0 {
		return 0, io.EOF
	}

	return totalRead, nil
}
