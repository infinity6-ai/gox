package iorunez

import (
	"bufio"
	"context"
	"io"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Options defines the parameters for the Read function, allowing for fine-grained
// control over the read operation.
type Options struct {
	// Min is the minimum number of runes to read. If the reader returns an EOF
	// before Min runes are read, Read returns io.ErrUnexpectedEOF.
	Min int
	// Max is the maximum number of runes to read. Read will not read more than
	// Max runes from the reader.
	Max int
	// Out is the buffer where the read data is stored. If Out has a non-zero
	// capacity, Read will append to it.
	Out []rune
	// Encoding is the encoding to use when reading from the reader. If nil,
	// UTF-8 is used.
	Encoding encoding.Encoding

	reader *bufio.Reader
}

// Read reads from r into opts.Out until the reader returns io.EOF or an error.
// It provides advanced control over the read operation through the Options struct.
//
// The function reads from r and appends the data to opts.Out. It respects the
// Min and Max fields of the opts parameter to control the amount of data read.
// The read operation can be cancelled via the context.
//
// To read a stream in chunks, the same Options instance must be used in subsequent calls.
// This is because the internal state of the decoder is associated with the Options object.
//
// Read returns the total number of runes read and an error if one occurred.
// If the number of runes read is less than opts.Min, it returns io.ErrUnexpectedEOF.
// If the reader is empty and no runes are read, it returns io.EOF.
func Read(ctx context.Context, r io.Reader, opts *Options) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	if opts.reader == nil {
		enc := opts.Encoding
		if enc == nil {
			enc = unicode.UTF8
		}
		opts.reader = bufio.NewReader(transform.NewReader(r, enc.NewDecoder()))
	}
	reader := opts.reader

	totalRead := 0
	var eofEncountered bool
	for {
		// check context in loop
		select {
		case <-ctx.Done():
			return totalRead, ctx.Err()
		default:
		}

		if opts.Max > 0 && totalRead >= opts.Max {
			break
		}

		rn, _, err := reader.ReadRune()
		if err == io.EOF {
			eofEncountered = true
			break
		}
		if err != nil {
			return totalRead, err
		}

		opts.Out = append(opts.Out, rn)
		totalRead++
	}

	if totalRead < opts.Min {
		return totalRead, io.ErrUnexpectedEOF
	}

	if eofEncountered && totalRead == 0 {
		return 0, io.EOF
	}

	return totalRead, nil
}
