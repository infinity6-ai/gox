package iorunez

import (
	"context"
	"io"
	"slices"
	"unicode/utf8"
)

const (
	DefaultGrowSize = 4096
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
	// GrowSize specifies the size by which the output buffer should grow when
	// more capacity is needed. If zero, a default size is used.
	GrowSize int
}

func (o *Options) grow(size int) []rune {
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
// Read returns the total number of runes read and an error if one occurred.
// If the number of runes read is less than opts.Min, it returns io.ErrUnexpectedEOF.
// If the reader is empty and no runes are read, it returns io.EOF.
func Read(ctx context.Context, r io.Reader, opts *Options) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	var eofEncountered bool
	var totalRead int

	// runeReader is a helper that reads runes from a byte stream.
	runeReader := func(br io.Reader) (rune, int, error) {
		// Buffer to read a single byte for UTF-8 decoding.
		// The size is 4 to handle the largest possible UTF-8 character.
		buf := make([]byte, 4)
		
		// Read 1 byte to start decoding.
		_, err := br.Read(buf[:1])
		if err != nil {
			return 0, 0, err
		}
		
		// If the byte is a single-byte character.
		if buf[0] < 128 {
			return rune(buf[0]), 1, nil
		}
		
		// Determine the number of additional bytes to read.
		extraBytes := 0
		if buf[0] >= 240 {
			extraBytes = 3
		} else if buf[0] >= 224 {
			extraBytes = 2
		} else if buf[0] >= 192 {
			extraBytes = 1
		}
		
		// Read the remaining bytes for the multi-byte character.
		_, err = io.ReadFull(br, buf[1:extraBytes+1])
		if err != nil {
			return 0, 0, err
		}
		
		// Decode the rune from the byte sequence.
		r, size := utf8.DecodeRune(buf[:extraBytes+1])
		return r, size, nil
	}

	for {
		select {
		case <-ctx.Done():
			return totalRead, ctx.Err()
		default:
		}

		if opts.Max > 0 && totalRead >= opts.Max {
			break
		}

		r, _, err := runeReader(r)
		if err != nil {
			if err == io.EOF {
				eofEncountered = true
				break
			}
			return totalRead, err
		}

		opts.Out = append(opts.Out, r)
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

