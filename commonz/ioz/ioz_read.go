package ioz

import "io"

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
