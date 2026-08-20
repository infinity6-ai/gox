package blobz

import (
	"bytes"
	"io"
	"reflect"
	"strings"
)

type Data interface {
	~string | ~[]byte
}

func ToBytes[T Data](val T) []byte {
	return New(val).Bytes()
}

func ToString[T Data](val T) string {
	return New(val).String()
}

type Blob interface {
	NewReader() io.Reader
	String() string
	Bytes() []byte
	IsString() bool
	IsBytes() bool
}

// wrapper is a non-generic struct holding either a string or []byte
type wrapper struct {
	data any
}

func (b *wrapper) NewReader() io.Reader {
	if s, ok := b.data.(string); ok {
		return strings.NewReader(s)
	}
	return bytes.NewReader(b.data.([]byte))
}

func (b *wrapper) String() string {
	if s, ok := b.data.(string); ok {
		return s
	}
	return string(b.data.([]byte))
}

func (b *wrapper) Bytes() []byte {
	if bs, ok := b.data.([]byte); ok {
		return bs
	}
	return []byte(b.data.(string))
}

func (b *wrapper) IsBytes() bool {
	_, ok := b.data.([]byte)
	return ok
}

func (b *wrapper) IsString() bool {
	_, ok := b.data.(string)
	return ok
}

func New[T Data](val T) Blob {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.String {
		return &wrapper{data: v.String()}
	}
	return &wrapper{data: v.Bytes()}
}
