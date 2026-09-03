package gobz

import (
	"io"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

// MustFormat is like Format but panics if an error occurs.
func MustFormat(v any) []byte {
	res, err := Format(v)
	errorz.Check(err)
	return res
}

// MustParse is like Parse but panics if an error occurs.
func MustParse[T any](data []byte, v T) T {
	res, err := Parse(data, v)
	errorz.Check(err)
	return res
}

// MustParseReader is like ParseReader but panics if an error occurs.
func MustParseReader[T any](r io.Reader, v T) T {
	res, err := ParseReader(r, v)
	errorz.Check(err)
	return res
}

// MustFormatWriter is like FormatWriter but panics if an error occurs.
func MustFormatWriter(w io.Writer, v any) {
	errorz.Check(FormatWriter(w, v))
}

// MustCopy is like Copy but panics if an error occurs.
func MustCopy[I any, O any](input I, output O) O {
	res, err := Copy(input, output)
	errorz.Check(err)
	return res
}

// MustClone is like Clone but panics if an error occurs.
func MustClone[T any](input T, output T) T {
	res, err := Clone(input, output)
	errorz.Check(err)
	return res
}
