package staticz

import "io"

type lineEncoder struct {
	w     io.Writer
	limit int
	count int
}

func (le *lineEncoder) Write(p []byte) (n int, err error) {
	for i, b := range p {
		if _, err := le.w.Write([]byte{b}); err != nil {
			return i, err
		}
		le.count++
		if le.count >= le.limit {
			if _, err := le.w.Write([]byte("\n")); err != nil {
				return i + 1, err
			}
			le.count = 0
		}
	}
	return len(p), nil
}
