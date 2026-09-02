package gzipz

import (
	"bytes"
	"compress/gzip"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/filez"
)

const GUNZIP_MAX_SIZE = 10 * 1025 * 1024

func Gzip[T blobz.Data](data T) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	defer gzipWriter.Close()
	_, err := gzipWriter.Write(blobz.New(data).Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to write gzip: %w", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to write gzip: %w", err)
	}
	return buf.Bytes(), nil
}

func MustGzip[T blobz.Data](data T) []byte {
	ret, err := Gzip(data)
	errorz.Check(err)
	return ret
}

func GunzipLimited(data []byte, maxSize int) (blobz.Blob, error) {
	buf := bytes.NewReader(data)
	gzipReader, err := gzip.NewReader(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to gunzip: %w", err)
	}
	defer gzipReader.Close()
	result := filez.ReadAllLimited(gzipReader, maxSize)
	return blobz.New(result), nil
}

func Gunzip(data []byte) (blobz.Blob, error) {
	return GunzipLimited(data, GUNZIP_MAX_SIZE)
}

func MustGunzip(data []byte) blobz.Blob {
	ret, err := Gunzip(data)
	errorz.Check(err)
	return ret
}
