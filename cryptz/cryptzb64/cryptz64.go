package cryptzb64

import (
	"encoding/base64"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
)

func newUrlEncoding() *base64.Encoding {
	return base64.URLEncoding.WithPadding(base64.NoPadding)
}

func newStdEncoding() *base64.Encoding {
	return base64.StdEncoding
}

func encode[T blobz.Data](enc *base64.Encoding, value T) blobz.Blob {
	data := blobz.ToBytes(value)
	buf := make([]byte, enc.EncodedLen(len(data)))
	enc.Encode(buf, data)
	return blobz.New(buf)
}

func decode[T blobz.Data](enc *base64.Encoding, value T) (blobz.Blob, error) {
	data := blobz.ToBytes(value)
	buf := make([]byte, enc.DecodedLen(len(data)))
	l, err := enc.Decode(buf, data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}
	return blobz.New(buf[:l]), nil
}

func UrlEncode[T blobz.Data](value T) blobz.Blob {
	return encode(newUrlEncoding(), value)
}

func UrlDecode[T blobz.Data](value T) (blobz.Blob, error) {
	return decode(newUrlEncoding(), value)
}

func StdEncode[T blobz.Data](value T) blobz.Blob {
	return encode(newStdEncoding(), value)
}

func StdDecode[T blobz.Data](value T) (blobz.Blob, error) {
	return decode(newStdEncoding(), value)
}
