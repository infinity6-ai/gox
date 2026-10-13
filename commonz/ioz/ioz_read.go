package ioz

import (
	"fmt"
	"io"
)

func ReadFull(r io.Reader, size int) ([]byte, error) {
	if r == nil {
		return nil, fmt.Errorf("reader cannot be nil")
	}
	if size <= 0 {
		return []byte{}, nil
	}
	buf := make([]byte, size)
	n := 0
	for n < size {
		nn, err := r.Read(buf[n:])
		n += nn
		if err != nil {
			if n >= size {
				return buf[:n], nil
			}
			return buf[:n], fmt.Errorf("failed to read full buffer: %w", err)
		}
	}
	return buf[:n], nil
}
