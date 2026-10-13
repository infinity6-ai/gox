package ioz

import (
	"io"
	"sync/atomic"
)

type ReaderCounter struct {
	io.Reader
	count atomic.Uint64
}

type WriterCounter struct {
	io.Reader
	count atomic.Uint64
}
