package cryptzrand

import (
	"crypto/rand"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

func Rand(size int) []byte {
	key := make([]byte, size)
	_, err := rand.Read(key)
	errorz.Check(err)
	return key
}
