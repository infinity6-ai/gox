package ioz

import (
	"bytes"
	"io"
)

type ReaderHeadTail struct {
	ReaderCounter
	HeadMaxSize int
	TailMaxSize int
	Head        bytes.Buffer
	Tail        bytes.Buffer // do not overlay with head
}

func NewReaderHeadTail(r io.Reader) *ReaderHeadTail {
	return &ReaderHeadTail{ReaderCounter: ReaderCounter{Reader: r}}
}

type WriterHeadTail struct {
	WriterCounter
	HeadMaxSize int
	TailMaxSize int
	Head        bytes.Buffer
	Tail        bytes.Buffer // do not overlay with head
}

func NewWriterHeadTail(w io.Writer) *WriterHeadTail {
	return &WriterHeadTail{WriterCounter: WriterCounter{Writer: w}}
}
