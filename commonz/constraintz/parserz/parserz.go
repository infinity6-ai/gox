package parserz

type ItemReader[T any] interface {
	ReadItem() (T, error)
}

type ItemWriter[T any] interface {
	WriteItem(item T) error
}

type ItemReaderWriter[T any] interface {
	ItemReader[T]
	ItemWriter[T]
}

type itemReaderWriter[T any] struct {
	read  func() (T, error)
	write func(item T) error
}

func (r *itemReaderWriter[T]) ReadItem() (T, error) {
	if r.read == nil {
		panic("ReadItem called on a write-only stream")
	}
	return r.read()
}

func (r *itemReaderWriter[T]) WriteItem(item T) error {
	if r.write == nil {
		panic("WriteItem called on a read-only stream")
	}
	return r.write(item)
}

func NewItemReaderWriter[T any](reader func() (T, error), writer func(item T) error) ItemReaderWriter[T] {
	return &itemReaderWriter[T]{
		read:  reader,
		write: writer,
	}
}
