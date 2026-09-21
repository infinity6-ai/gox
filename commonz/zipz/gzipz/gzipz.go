// Package gzipz provides helper utilities for compressing and decompressing
// data using the gzip format.
package gzipz

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/filez"
)

// GUNZIP_MAX_SIZE is the default maximum uncompressed size (10MB)
// allowed during decompression to prevent decompression bomb vulnerabilities.
const GUNZIP_MAX_SIZE = 10 * 1025 * 1024

// Gzip compresses the provided data (which can be a string or a byte slice)
// using gzip compression and returns the compressed bytes.
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

// MustGzip compresses the provided data using gzip compression and returns
// the compressed bytes. It panics if any error occurs.
func MustGzip[T blobz.Data](data T) []byte {
	ret, err := Gzip(data)
	errorz.Check(err)
	return ret
}

// GunzipLimited decompresses the provided gzip-compressed data up to the
// specified maximum uncompressed size (in bytes). It panics if the
// decompressed size exceeds maxSize, or returns an error if decompression fails.
func GunzipLimited(data []byte, maxSize int) (blobz.Blob, error) {
	buf := bytes.NewReader(data)
	gzipReader, err := gzip.NewReader(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to gunzip: %w", err)
	}
	defer gzipReader.Close()
	result := filez.ReadAllLimited(gzipReader, maxSize)
	return result, nil
}

// Gunzip decompresses the provided gzip-compressed data using the default
// GUNZIP_MAX_SIZE limit (10MB). It panics if the decompressed size exceeds
// GUNZIP_MAX_SIZE, or returns an error if decompression fails.
func Gunzip(data []byte) (blobz.Blob, error) {
	return GunzipLimited(data, GUNZIP_MAX_SIZE)
}

// MustGunzip decompresses the provided gzip-compressed data and returns the
// uncompressed Blob. It panics if decompression fails or if the decompressed
// size exceeds GUNZIP_MAX_SIZE.
func MustGunzip(data []byte) blobz.Blob {
	ret, err := Gunzip(data)
	errorz.Check(err)
	return ret
}

// MustReader wraps the provided io.Reader with a gzip.Reader. It panics
// if the reader initialization fails. The returned io.ReadCloser should be
// closed by the caller when done.
func MustReader(r io.Reader) io.ReadCloser {
	ret, err := gzip.NewReader(r)
	errorz.Check(err)
	return ret
}
