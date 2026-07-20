package iobytez

import (
	"context"
	"io"
)

// Options sets the parameters for the Read function.
type Options struct {
	// Out is the buffer to which the data will be appended.
	// If nil, a new buffer will be allocated with 4K capacity.
	Out []byte
	// Min is the minimum number of bytes to read.
	// If the reader returns an EOF before Min bytes are read, io.ErrUnexpectedEOF is returned.
	Min int
	// Max is the maximum number of bytes to read.
	// The read will stop once Max bytes are read.
	Max int
}

// Read reads from r into opts.Out until at least opts.Min bytes have been read or opts.Max bytes have been read.
// It uses a timeout if specified in opts.Timeout.
func Read(ctx context.Context, r io.Reader, opts *Options) error {
	panic("implement it")
}
