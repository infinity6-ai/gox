package cryptzb64

import (
	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/encz/enczb64"
)

func UrlEncode[T blobz.Data](value T) blobz.Blob {
	return enczb64.UrlEncode(value)
}

func UrlDecode[T blobz.Data](value T) (blobz.Blob, error) {
	return enczb64.UrlDecode(value)
}

func StdEncode[T blobz.Data](value T) blobz.Blob {
	return enczb64.StdEncode(value)
}

func StdDecode[T blobz.Data](value T) (blobz.Blob, error) {
	return enczb64.StdDecode(value)
}
