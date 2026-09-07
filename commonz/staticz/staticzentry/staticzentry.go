package staticzentry

import "io"

type Entry interface {
	Name() string
	Size() int64
	Open() (io.ReadCloser, error)
}

type entry struct {
	name string
	size int64
	open func() (io.ReadCloser, error)
}

func NewEntry(name string, size int64, open func() (io.ReadCloser, error)) Entry {
	return &entry{
		name: name,
		size: size,
		open: open,
	}
}

func (e *entry) Size() int64 {
	return e.size
}

func (e *entry) Name() string {
	return e.name
}

func (e *entry) Open() (io.ReadCloser, error) {
	return e.open()
}
