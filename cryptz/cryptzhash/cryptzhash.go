package cryptzhash

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
)

const SHA256Size = sha256.Size

func SHA256Data[T blobz.Data](val T) ([]byte, error) {
	return SHA256(blobz.New(val).NewReader())
}

func SHA256(reader io.Reader) ([]byte, error) {
	hasher := sha256.New()
	_, err := io.Copy(hasher, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to copy: %w", err)
	}
	return hasher.Sum(nil), nil
}

const MD5Size = md5.Size

func MD5Data[T blobz.Data](val T) ([]byte, error) {
	return MD5(blobz.New(val).NewReader())
}

func MD5(reader io.Reader) ([]byte, error) {
	hasher := md5.New()
	_, err := io.Copy(hasher, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to copy: %w", err)
	}
	return hasher.Sum(nil), nil
}
