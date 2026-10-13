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

func NewReaderHeadTail(r io.Reader, headMaxSize int, tailMaxSize int) *ReaderHeadTail {
	return &ReaderHeadTail{
		ReaderCounter: ReaderCounter{Reader: r},
		HeadMaxSize:   headMaxSize,
		TailMaxSize:   tailMaxSize,
	}
}

type WriterHeadTail struct {
	WriterCounter
	HeadMaxSize int
	TailMaxSize int
	Head        bytes.Buffer
	Tail        bytes.Buffer // do not overlay with head
}

func NewWriterHeadTail(w io.Writer, headMaxSize int, tailMaxSize int) *WriterHeadTail {
	return &WriterHeadTail{
		WriterCounter: WriterCounter{Writer: w},
		HeadMaxSize:   headMaxSize,
		TailMaxSize:   tailMaxSize,
	}
}
