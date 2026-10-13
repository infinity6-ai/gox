package ioz

import (
	"io"
)

type ReaderCounter struct {
	io.Reader
	count uint64
}

func (r *ReaderCounter) Count() uint64 {
	return r.count
}

type WriterCounter struct {
	io.Writer
	count uint64
}

func (w *WriterCounter) Count() uint64 {
	return w.count
}
