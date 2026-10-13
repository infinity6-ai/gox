package ioz

import (
	"bytes"
	"io"
)

type ReaderHeadTail struct {
	io.Reader
	HeadMaxSize int
	TailMaxSize int
	Head        bytes.Buffer
	Tail        bytes.Buffer // do not overlay with head
}

type WriterHeadTail struct {
	io.Writer
	HeadMaxSize int
	TailMaxSize int
	Head        bytes.Buffer
	Tail        bytes.Buffer // do not overlay with head
}
