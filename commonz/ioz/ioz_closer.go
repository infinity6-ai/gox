package ioz

import "io"

type closer struct {
	io.Closer
	f func() error
}

func (c *closer) Close() error {
	if c.f == nil {
		return nil
	}
	return c.f()
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

func ReadFull(r io.Reader, size int) ([]byte, error) {
	buf := make([]byte, size)
	n := 0
	for {
		if n >= size {
			return buf[:n], nil
		}
		nn, err := r.Read(buf[n:])
		n += nn
		if err != nil {
			return buf[:n], err
		}
	}
}
