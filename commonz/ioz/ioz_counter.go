package ioz

import (
	"io"
	"sync/atomic"
)

type ReaderCounter struct {
	io.Reader
	count atomic.Uint64
}

func (r *ReaderCounter) Count() uint64 {
	return r.count.Load()
}

type WriterCounter struct {
	io.Writer
	count atomic.Uint64
}

func (w *WriterCounter) Count() uint64 {
	return w.count.Load()
}
