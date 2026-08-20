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
}

// wrapper is a non-generic struct holding either a string or []byte
type wrapper struct {
	data any
}

func (b *wrapper) NewReader() io.Reader {
	if s, ok := b.data.(string); ok {
		return strings.NewReader(s)
	}
	if bs, ok := b.data.([]byte); ok {
		return bytes.NewReader(bs)
	}

	// Fallback for aliased types
	rv := reflect.ValueOf(b.data)
	if rv.Kind() == reflect.String {
		return strings.NewReader(rv.String())
	}
	return bytes.NewReader(rv.Bytes())
}

func (b *wrapper) String() string {
	if s, ok := b.data.(string); ok {
		return s
	}
	if bs, ok := b.data.([]byte); ok {
		return string(bs)
	}
	// Fallback for aliased types
	rv := reflect.ValueOf(b.data)
	if rv.Kind() == reflect.String {
		return rv.String()
	}
	return string(rv.Bytes())
}

func (b *wrapper) Bytes() []byte {
	if bs, ok := b.data.([]byte); ok {
		return bs
	}
	if s, ok := b.data.(string); ok {
		return []byte(s)
	}
	// Fallback for aliased types
	rv := reflect.ValueOf(b.data)
	if rv.Kind() == reflect.String {
		return []byte(rv.String())
	}
	return rv.Bytes()
}

func (b *wrapper) IsString() bool {
	if _, ok := b.data.(string); ok {
		return true
	}
	if _, ok := b.data.([]byte); ok {
		return false
	}
	// Fallback for aliased types
	return reflect.ValueOf(b.data).Kind() == reflect.String
}

func New[T Data](val T) Blob {
	return &wrapper{data: val}
}
