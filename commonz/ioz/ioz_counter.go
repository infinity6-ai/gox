package ioz

import (
	"fmt"
	"io"
)

type Countable interface {
	GetCount() uint64
}

type ReaderCounter struct {
	io.Reader
	count uint64
}

func NewReaderCounter(r io.Reader) *ReaderCounter {
	return &ReaderCounter{Reader: r}
}

func (r *ReaderCounter) Read(p []byte) (n int, err error) {
	if r == nil || r.Reader == nil {
		return 0, io.EOF
	}
	n, err = r.Reader.Read(p)
	r.count += uint64(n)
	if err != nil && err != io.EOF {
		return n, fmt.Errorf("failed to read: %w", err)
	}
	return n, err
}

func (r *ReaderCounter) GetCount() uint64 {
	if r == nil {
		return 0
	}
	return r.count
}

type WriterCounter struct {
	io.Writer
	count uint64
}

func NewWriterCounter(w io.Writer) *WriterCounter {
	return &WriterCounter{Writer: w}
}

func (w *WriterCounter) Write(p []byte) (n int, err error) {
	if w == nil {
		return 0, nil
	}
	if w.Writer == nil {
		w.count += uint64(len(p))
		return len(p), nil
	}
	n, err = w.Writer.Write(p)
	w.count += uint64(n)
	if err != nil {
		return n, fmt.Errorf("failed to write: %w", err)
	}
	return n, nil
}

func (w *WriterCounter) GetCount() uint64 {
	if w == nil {
		return 0
	}
	return w.count
}
