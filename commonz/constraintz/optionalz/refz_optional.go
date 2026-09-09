package optionalz

type Optional[T any] struct {
	value   T
	present bool
}

func New[T any](value T, present bool) Optional[T] {
	return Optional[T]{
		value:   value,
		present: present,
	}
}

func Present[T any](value T) Optional[T] {
	return Optional[T]{
		value:   value,
		present: true,
	}
}

func Empty[T any]() Optional[T] {
	return Optional[T]{
		present: false,
	}
}

func (o Optional[T]) IsPresent() bool {
	return o.present
}

func (o Optional[T]) Get() (T, bool) {
	return o.value, o.present
}

func (o Optional[T]) Must() T {
	if !o.present {
		panic("value not present")
	}
	return o.value
}
