package parserz

import (
	"github.com/infinity6-ai/gox/commonz/errorz"
)

type ItemReader[T any] interface {
	ReadItemInto(item T) (T, error)
	MustReadItemInto(item T) T
}

type ItemWriter[T any] interface {
	WriteItem(item T) error
	MustWriteItem(item T)
}

type ItemReaderWriter[T any] interface {
	ItemReader[T]
	ItemWriter[T]
}

type MustItemWriter[T any] interface {
	MustWriteItem(item T)
}

type itemReaderWriter[T any] struct {
	read  func() (T, error)
	write func(item T) error
}

func (r *itemReaderWriter[T]) ReadItemInto(item T) (T, error) {
	if r.read == nil {
		panic("ReadItemInto called on a write-only stream")
	}
	return r.read()
}

func (r *itemReaderWriter[T]) MustReadItemInto(item T) T {
	res, err := r.ReadItemInto(item)
	errorz.Check(err)
	return res
}

func (r *itemReaderWriter[T]) WriteItem(item T) error {
	if r.write == nil {
		panic("WriteItem called on a read-only stream")
	}
	return r.write(item)
}

func (r *itemReaderWriter[T]) MustWriteItem(item T) {
	errorz.Check(r.WriteItem(item))
}

func NewItemReaderWriter[T any](reader func() (T, error), writer func(item T) error) ItemReaderWriter[T] {
	return &itemReaderWriter[T]{
		read:  reader,
		write: writer,
	}
}

func NewMustItemWriter[T any](writer func(item T) error) MustItemWriter[T] {
	return &itemReaderWriter[T]{
		write: writer,
	}
}
