package iobytez

import (
	"context"
	"io"
	"time"
)

// Options sets the parameters for the Read function.
type Options struct {
	// Out is the buffer to which the data will be read.
	// If nil, a new buffer will be allocated.
	Out []byte
	// Min is the minimum number of bytes to read.
	// If the reader returns an EOF before Min bytes are read, io.ErrUnexpectedEOF is returned.
	Min int
	// Max is the maximum number of bytes to read.
	// The read will stop once Max bytes are read.
	Max int
	// Timeout is the maximum time to wait for the read to complete.
	// If the timeout is reached before Min bytes are read, a context error is returned.
	// If the timeout is reached after Min bytes are read, the function returns successfully with the bytes read so far.
	Timeout time.Duration
}

// Read reads from r into opts.Out until at least opts.Min bytes have been read or opts.Max bytes have been read.
// It uses a timeout if specified in opts.Timeout.
func Read(ctx context.Context, r io.Reader, opts *Options) error {
	if opts.Out == nil {
		opts.Out = make([]byte, 0, 4096)
	}
	if opts.Max <= 0 {
		opts.Max = cap(opts.Out)
	}
	if opts.Min > opts.Max {
		opts.Min = opts.Max
	}
	
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	buf := make([]byte, 4096)
	bytesRead := 0

	for bytesRead < opts.Max {
		type readResult struct {
			n   int
			err error
		}
		readCh := make(chan readResult, 1)

		go func() {
			// Adjust buffer size for the read.
			readSize := len(buf)
			if needed := opts.Max - bytesRead; needed < readSize {
				readSize = needed
			}
			n, err := r.Read(buf[:readSize])
			readCh <- readResult{n, err}
		}()

		select {
		case <-ctx.Done():
			if bytesRead >= opts.Min {
				return nil
			}
			return ctx.Err()
		case res := <-readCh:
			if res.n > 0 {
				opts.Out = append(opts.Out, buf[:res.n]...)
				bytesRead += res.n
			}
			if res.err != nil {
				if res.err == io.EOF {
					if bytesRead >= opts.Min {
						return nil
					}
					if bytesRead > 0 {
						return io.ErrUnexpectedEOF
					}
					return io.EOF
				}
				return res.err
			}

			if bytesRead >= opts.Min && opts.Timeout == 0 {
				return nil
			}
		}
	}
	return nil
}
