package jsonz

import (
	"io"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
)

func LineParseReader[S ~[]E, E any](r io.Reader, v *S) error {
	panic("implement it")
}

func LineParseInto[S ~[]E, E any, I blobz.Data](data I, v *S) error {
	panic("implement it")
}

func LineFormatWriter[S ~[]E, E any](w io.Writer, v S) error {
	panic("implement it")
}

func LineFormatReadCloser[S ~[]E, E any](v S) io.ReadCloser {
	panic("implement it")
}

func LineFormatReader[S ~[]E, E any](v S, fn func(r io.Reader) error) error {
	panic("implement it")
}

func LineFormat[S ~[]E, E any](v S) (blobz.Blob, error) {
	panic("implement it")
}

func LineFormatBytes[S ~[]E, E any](v S) ([]byte, error) {
	ret, err := LineFormat(v)
	if err != nil {
		return nil, err
	}
	return ret.Bytes(), nil
}
