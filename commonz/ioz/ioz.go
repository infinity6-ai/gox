package ioz

import "io"

type closer struct {
	io.Closer
	f func() error
}

func (me *closer) Close() error {
	return me.f()
}

func Closer(f func() error) io.Closer {
	return &closer{f: f}
}

func CloserV(f func()) io.Closer {
	return &closer{f: func() error {
		f()
		return nil
	}}
}
