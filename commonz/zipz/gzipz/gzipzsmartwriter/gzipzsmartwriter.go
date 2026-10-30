// Package gzipzsmartwriter provides a smart gzip writer that buffers initial data to decide
// whether gzip compression is worthwhile, passing through uncompressed data if not.
package gzipzsmartwriter

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// MinHead is the default minimum viable head size in bytes (1024)
// used to evaluate whether compression is worthwhile.
const MinHead = 1024

const minHead = MinHead

// Opts configures the behavior of smart gzip compression.
type Opts struct {
	// Head specifies how many bytes are buffered to decide whether compression is worthwhile.
	// If set to less than MinHead (1024 bytes), MinHead is used.
	Head int

	// Decision is called before any bytes are written to the underlying writer,
	// indicating whether gzip compression will be applied.
	Decision func(compress bool)
}

type smartGzipWriter struct {
	w            io.Writer
	opts         Opts
	headBuf      bytes.Buffer
	decisionMade bool
	compressing  bool
	gzWriter     *gzip.Writer
	closed       bool
}

// New creates a new smart gzip io.WriteCloser. It buffers up to opts.Head bytes (minimum 1024)
// to evaluate whether gzip compression reduces payload size. If compression is beneficial,
// it streams compressed gzip bytes; otherwise, it passes raw bytes through.
// The Decision callback (if configured) is guaranteed to be invoked before any bytes are written
// to the underlying writer.
// Closing the returned io.WriteCloser flushes any pending data but never closes the underlying writer.
func New(w io.Writer, opts Opts) io.WriteCloser {
	if opts.Head < minHead {
		opts.Head = minHead
	}

	return &smartGzipWriter{
		w:    w,
		opts: opts,
	}
}

// isWorthGzip tests whether compressing data with gzip results in fewer bytes than the original data.
func isWorthGzip(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write(data)
	if err != nil {
		return false
	}
	if err := gw.Close(); err != nil {
		return false
	}
	return buf.Len() < len(data)
}

func (s *smartGzipWriter) makeDecision(data []byte) {
	s.decisionMade = true
	s.compressing = isWorthGzip(data)
	if s.opts.Decision != nil {
		s.opts.Decision(s.compressing)
	}
	if s.compressing {
		s.gzWriter = gzip.NewWriter(s.w)
	}
}

// Write writes data to the smart gzip writer. It buffers initial data until opts.Head
// bytes are accumulated to make a compression decision, then flushes and streams data.
func (s *smartGzipWriter) Write(p []byte) (int, error) {
	if s.closed {
		return 0, fmt.Errorf("gzipzsmartwriter: write to closed writer")
	}

	if s.decisionMade {
		if s.compressing {
			n, err := s.gzWriter.Write(p)
			if err != nil {
				return n, fmt.Errorf("failed to write to gzip writer: %w", err)
			}
			return n, nil
		}
		n, err := s.w.Write(p)
		if err != nil {
			return n, fmt.Errorf("failed to write to underlying writer: %w", err)
		}
		return n, nil
	}

	needed := s.opts.Head - s.headBuf.Len()
	if len(p) < needed {
		n, err := s.headBuf.Write(p)
		if err != nil {
			return n, fmt.Errorf("failed to buffer head data: %w", err)
		}
		return n, nil
	}

	// We now have at least opts.Head bytes to make a decision
	_, err := s.headBuf.Write(p[:needed])
	if err != nil {
		return 0, fmt.Errorf("failed to buffer head data: %w", err)
	}

	s.makeDecision(s.headBuf.Bytes())

	if s.compressing {
		_, err = s.gzWriter.Write(s.headBuf.Bytes())
		if err != nil {
			return 0, fmt.Errorf("failed to write head to gzip writer: %w", err)
		}
		s.headBuf.Reset()

		if len(p) > needed {
			n, err := s.gzWriter.Write(p[needed:])
			if err != nil {
				return needed + n, fmt.Errorf("failed to write body to gzip writer: %w", err)
			}
		}
		return len(p), nil
	}

	// Not compressing: pass through head and remaining bytes
	_, err = s.w.Write(s.headBuf.Bytes())
	if err != nil {
		return 0, fmt.Errorf("failed to write head to underlying writer: %w", err)
	}
	s.headBuf.Reset()

	if len(p) > needed {
		n, err := s.w.Write(p[needed:])
		if err != nil {
			return needed + n, fmt.Errorf("failed to write body to underlying writer: %w", err)
		}
	}
	return len(p), nil
}

// Close flushes any remaining buffered data, makes a decision if none was made yet,
// finalizes gzip compression if applicable, and never closes the underlying writer.
func (s *smartGzipWriter) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true

	if !s.decisionMade {
		if s.headBuf.Len() == 0 {
			if s.opts.Decision != nil {
				s.opts.Decision(false)
			}
			s.decisionMade = true
			return nil
		}

		s.makeDecision(s.headBuf.Bytes())
		if s.compressing {
			_, err := s.gzWriter.Write(s.headBuf.Bytes())
			if err != nil {
				return fmt.Errorf("failed to write head to gzip writer on close: %w", err)
			}
			s.headBuf.Reset()
			if err := s.gzWriter.Close(); err != nil {
				return fmt.Errorf("failed to close gzip writer: %w", err)
			}
			return nil
		}

		_, err := s.w.Write(s.headBuf.Bytes())
		if err != nil {
			return fmt.Errorf("failed to write head to underlying writer on close: %w", err)
		}
		s.headBuf.Reset()
		return nil
	}

	if s.compressing && s.gzWriter != nil {
		if err := s.gzWriter.Close(); err != nil {
			return fmt.Errorf("failed to close gzip writer: %w", err)
		}
	}

	// The underlying writer is intentionally never closed
	return nil
}

// Flush flushes buffered data to the underlying writer, forcing a decision if one
// has not been made yet.
func (s *smartGzipWriter) Flush() error {
	if s.closed {
		return fmt.Errorf("gzipzsmartwriter: flush on closed writer")
	}

	if !s.decisionMade {
		if s.headBuf.Len() == 0 {
			return nil
		}
		s.makeDecision(s.headBuf.Bytes())
		if s.compressing {
			_, err := s.gzWriter.Write(s.headBuf.Bytes())
			if err != nil {
				return fmt.Errorf("failed to write head to gzip writer on flush: %w", err)
			}
			s.headBuf.Reset()
			if err := s.gzWriter.Flush(); err != nil {
				return fmt.Errorf("failed to flush gzip writer: %w", err)
			}
		} else {
			_, err := s.w.Write(s.headBuf.Bytes())
			if err != nil {
				return fmt.Errorf("failed to write head to underlying writer on flush: %w", err)
			}
			s.headBuf.Reset()
		}
	} else if s.compressing && s.gzWriter != nil {
		if err := s.gzWriter.Flush(); err != nil {
			return fmt.Errorf("failed to flush gzip writer: %w", err)
		}
	}

	if flusher, ok := s.w.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			return fmt.Errorf("failed to flush underlying writer: %w", err)
		}
	} else if flusher, ok := s.w.(interface{ Flush() }); ok {
		flusher.Flush()
	}

	return nil
}

// Reset resets the smart gzip writer to write to a new underlying writer,
// allowing reuse without re-allocating.
func (s *smartGzipWriter) Reset(w io.Writer) {
	s.w = w
	s.headBuf.Reset()
	s.decisionMade = false
	s.compressing = false
	s.gzWriter = nil
	s.closed = false
}
