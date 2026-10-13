package ioz

import (
	"bytes"
	"io"
)

type HeadTail interface {
	Countable
	HeadBytes() []byte
	TailBytes() []byte
}

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

func (r *ReaderHeadTail) GetHead() bytes.Buffer {
	return r.Head
}

func (r *ReaderHeadTail) GetTail() bytes.Buffer {
	return r.Tail
}

func (r *ReaderHeadTail) Read(p []byte) (n int, err error) {
	n, err = r.ReaderCounter.Read(p)
	if n > 0 {
		appendHeadTail(&r.Head, r.HeadMaxSize, &r.Tail, r.TailMaxSize, p[:n])
	}
	return n, err
}

func (r *ReaderHeadTail) HeadString() string {
	if r == nil {
		return ""
	}
	return r.Head.String()
}

func (r *ReaderHeadTail) TailString() string {
	if r == nil {
		return ""
	}
	return r.Tail.String()
}

func (r *ReaderHeadTail) HeadBytes() []byte {
	if r == nil || r.Head.Len() == 0 {
		return []byte{}
	}
	return r.Head.Bytes()
}

func (r *ReaderHeadTail) TailBytes() []byte {
	if r == nil || r.Tail.Len() == 0 {
		return []byte{}
	}
	return r.Tail.Bytes()
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

func (w *WriterHeadTail) GetHead() bytes.Buffer {
	return w.Head
}

func (w *WriterHeadTail) GetTail() bytes.Buffer {
	return w.Tail
}

func (w *WriterHeadTail) Write(p []byte) (n int, err error) {
	n, err = w.WriterCounter.Write(p)
	if n > 0 {
		appendHeadTail(&w.Head, w.HeadMaxSize, &w.Tail, w.TailMaxSize, p[:n])
	}
	return n, err
}

func (w *WriterHeadTail) HeadString() string {
	if w == nil {
		return ""
	}
	return w.Head.String()
}

func (w *WriterHeadTail) TailString() string {
	if w == nil {
		return ""
	}
	return w.Tail.String()
}

func (w *WriterHeadTail) HeadBytes() []byte {
	if w == nil || w.Head.Len() == 0 {
		return []byte{}
	}
	return w.Head.Bytes()
}

func (w *WriterHeadTail) TailBytes() []byte {
	if w == nil || w.Tail.Len() == 0 {
		return []byte{}
	}
	return w.Tail.Bytes()
}

func appendHeadTail(head *bytes.Buffer, headMaxSize int, tail *bytes.Buffer, tailMaxSize int, p []byte) {
	if headMaxSize > 0 && head.Len() < headMaxSize {
		take := headMaxSize - head.Len()
		if take > len(p) {
			take = len(p)
		}
		head.Write(p[:take])
		p = p[take:]
	}

	if tailMaxSize <= 0 || len(p) == 0 {
		return
	}

	if len(p) >= tailMaxSize {
		tail.Reset()
		tail.Write(p[len(p)-tailMaxSize:])
	} else if tail.Len()+len(p) > tailMaxSize {
		needFromCurrent := tailMaxSize - len(p)
		kept := append([]byte(nil), tail.Bytes()[tail.Len()-needFromCurrent:]...)
		tail.Reset()
		tail.Write(kept)
		tail.Write(p)
	} else {
		tail.Write(p)
	}
}
