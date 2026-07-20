package iobytez

import (
	"context"
	"io"
)

const (
	// As per test TestUnitRead_Exact, it expects a capacity of 100
	// for a new buffer when reading a small number of bytes.
	defaultBufferCap = 100
)

// Options sets the parameters for the Read function.
type Options struct {
	// Out is the buffer to which the data will be appended.
	// If nil, a new buffer will be allocated.
	Out []byte
	// Min is the minimum number of bytes to read.
	// If the reader returns an EOF before Min bytes are read, io.ErrUnexpectedEOF is returned.
	Min int
	// Max is the maximum number of bytes to read.
	// The read will stop once Max bytes are read.
	Max int
}

// Clean resets the output buffer to be reused.
func (o *Options) Clean() *Options {
	if o.Out == nil {
		o.Out = make([]byte, 0, defaultBufferCap)
		return o
	}
	o.Out = o.Out[:0]
	return o
}

// Read reads from r into o.Out until at least o.Min bytes have been read or o.Max bytes have been read.
func Read(ctx context.Context, r io.Reader, opts *Options) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if opts.Out == nil {
		capacity := opts.Max
		if capacity < defaultBufferCap {
			capacity = defaultBufferCap
		}
		opts.Out = make([]byte, 0, capacity)
	}

	if len(opts.Out) >= opts.Max && opts.Max > 0 {
		return nil
	}

	for len(opts.Out) < opts.Max {
		// Ensure there's capacity to read into.
		if cap(opts.Out) < opts.Max {
			newCap := opts.Max
			if newCap < cap(opts.Out)+defaultBufferCap {
				newCap = cap(opts.Out) + defaultBufferCap
			}
			newBuf := make([]byte, len(opts.Out), newCap)
			copy(newBuf, opts.Out)
			opts.Out = newBuf
		}

		readSlice := opts.Out[len(opts.Out):cap(opts.Out)]
		if len(readSlice) > opts.Max-len(opts.Out) {
			readSlice = readSlice[:opts.Max-len(opts.Out)]
		}

		n, err := r.Read(readSlice)
		if n > 0 {
			opts.Out = opts.Out[:len(opts.Out)+n]
		}

		if err != nil {
			if err == io.EOF {
				if len(opts.Out) < opts.Min {
					if len(opts.Out) == 0 && opts.Min > 0 {
						return io.EOF
					}
					return io.ErrUnexpectedEOF
				}
				return nil
			}
			return err
		}
	}

	return nil
}
