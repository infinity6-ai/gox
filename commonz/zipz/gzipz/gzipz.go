package gzipz

import (
	"bytes"
	"compress/gzip"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"go.code.infinity6.ai/platform/util"
)

const GUNZIP_MAX_SIZE = 10 * 1025 * 1024

func Gzip(data blobz.Blob) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	defer gzipWriter.Close()
	_, err := gzipWriter.Write(data.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to write gzip: %w", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to write gzip: %w", err)
	}
	return buf.Bytes(), nil
}

func GunzipLimited(data []byte, maxSize int) (blobz.Blob, error) {
	buf := bytes.NewReader(data)
	gzipReader, err := gzip.NewReader(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to gunzip: %w", err)
	}
	defer gzipReader.Close()
	result := util.ReadAllLimited(gzipReader, maxSize)
	return blobz.New(result), nil
}

func Gunzip(data []byte) (blobz.Blob, error) {
	return GunzipLimited(data, GUNZIP_MAX_SIZE)
}
