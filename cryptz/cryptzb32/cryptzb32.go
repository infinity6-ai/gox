package cryptzb32

import (
	"encoding/base32"
	"io"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/errorz"
)

func newEncoding() *base32.Encoding {
	return base32.NewEncoding("0123456789abcdefghijklmnopqrstuv").WithPadding(base32.NoPadding)
}

func NewEncoder(w io.Writer) io.WriteCloser {
	enc := newEncoding()
	return base32.NewEncoder(enc, w)
}

func NewDecoder(r io.Reader) io.Reader {
	enc := newEncoding()
	return base32.NewDecoder(enc, r)
}

func encode[T blobz.Data](enc *base32.Encoding, value T) blobz.Blob {
	data := blobz.ToBytes(value)
	buf := make([]byte, enc.EncodedLen(len(data)))
	enc.Encode(buf, data)
	return blobz.New(buf)
}

func decode[T blobz.Data](enc *base32.Encoding, value T) blobz.Blob {
	data := blobz.ToBytes(value)
	buf := make([]byte, enc.DecodedLen(len(data)))
	l, err := enc.Decode(buf, data)
	errorz.Check(err)
	return blobz.New(buf[:l])
}

func Encode[T blobz.Data](value T) blobz.Blob {
	enc := newEncoding()
	return encode(enc, value)
}

func Decode[T blobz.Data](value T) blobz.Blob {
	enc := newEncoding()
	return decode(enc, value)
}

func EncodeCopy(out io.Writer, in io.Reader) int64 {
	w := NewEncoder(out)
	defer w.Close()
	ret, err := io.Copy(w, in)
	errorz.Check(err)
	return ret
}

func DecodeCopy(out io.Writer, in io.Reader) int64 {
	r := NewDecoder(in)
	ret, err := io.Copy(out, r)
	errorz.Check(err)
	return ret
}
