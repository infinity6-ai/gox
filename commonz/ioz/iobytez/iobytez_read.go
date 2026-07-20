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
}

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
