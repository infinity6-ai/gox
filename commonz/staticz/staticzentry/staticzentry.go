package staticzentry

import (
	"io"

	"github.com/infinity6-ai/gox/commonz/pathz"
)

type Entry interface {
	Name() *pathz.Path
	Size() int64
	Open() (io.ReadCloser, error)
}

type entry struct {
	name *pathz.Path
	size int64
	open func() (io.ReadCloser, error)
}

func NewEntry(name *pathz.Path, size int64, open func() (io.ReadCloser, error)) Entry {
	name.Check(pathz.ValidateOptions{
		Absolute:    new(false),
		MaxParents:  new(0),
		Wildchar:    false,
		EndingSlash: new(false),
		Empty:       new(false),
	})
	return &entry{
		name: name,
		size: size,
		open: open,
	}
}

func (e *entry) Size() int64 {
	return e.size
}

func (e *entry) Name() *pathz.Path {
	return e.name
}

func (e *entry) Open() (io.ReadCloser, error) {
	return e.open()
}
