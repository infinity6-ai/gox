package parserz

type ItemReader[T any] interface {
	ReadItem() (*T, error)
}

type ItemWriter[T any] interface {
	WriteItem(item *T) error
}

type ItemReaderWriter[T any] interface {
	ItemReader[T]
	ItemWriter[T]
}
