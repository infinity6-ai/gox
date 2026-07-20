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
// Min and Max are relative to the number of bytes read in this call, not the total size of Out.
func Read(ctx context.Context, r io.Reader, opts *Options) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	startLen := len(opts.Out)
	if opts.Out == nil {
		capacity := opts.Max
		if capacity < defaultBufferCap {
			capacity = defaultBufferCap
		}
		opts.Out = make([]byte, 0, capacity)
	}

	if opts.Max <= 0 {
		if opts.Min > 0 {
			return io.ErrUnexpectedEOF
		}
		return nil
	}

	var lastErr error
	for {
		bytesRead := len(opts.Out) - startLen
		if bytesRead >= opts.Max {
			break
		}

		neededCap := startLen + opts.Max
		if cap(opts.Out) < neededCap {
			newCap := neededCap
			if newCap < cap(opts.Out)+defaultBufferCap {
				newCap = cap(opts.Out) + defaultBufferCap
			}
			newBuf := make([]byte, len(opts.Out), newCap)
			copy(newBuf, opts.Out)
			opts.Out = newBuf
		}

		readSlice := opts.Out[len(opts.Out):cap(opts.Out)]
		if len(readSlice) > (opts.Max - bytesRead) {
			readSlice = readSlice[:(opts.Max - bytesRead)]
		}

		n, err := r.Read(readSlice)
		if n > 0 {
			opts.Out = opts.Out[:len(opts.Out)+n]
		}

		if err != nil {
			lastErr = err
			break
		}
		if n == 0 {
			break
		}
	}

	bytesRead := len(opts.Out) - startLen
	if lastErr == io.EOF {
		if bytesRead < opts.Min {
			if startLen == 0 && bytesRead == 0 && opts.Min > 0 {
				return io.EOF
			}
			return io.ErrUnexpectedEOF
		}
		if bytesRead == 0 {
			return io.EOF
		}
		return nil
	}

	if bytesRead < opts.Min {
		return io.ErrUnexpectedEOF
	}

	return nil
}
