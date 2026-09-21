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
	return result, nil
}

func Gunzip(data []byte) (blobz.Blob, error) {
	return GunzipLimited(data, GUNZIP_MAX_SIZE)
}

func MustGunzip(data []byte) blobz.Blob {
	ret, err := Gunzip(data)
	errorz.Check(err)
	return ret
}

// Reader returns an io.ReadCloser that decompresses gzip-compressed data read from r.
// The underlying gzip.Reader is initialized lazily upon the first Read call.
func Reader(r io.Reader) io.ReadCloser {
	return &lazyReader{r: r}
}

type lazyReader struct {
	r   io.Reader
	gr  *gzip.Reader
	err error
}

func (lr *lazyReader) Read(p []byte) (n int, err error) {
	if lr.err != nil {
		return 0, lr.err
	}
	if lr.gr == nil {
		lr.gr, lr.err = gzip.NewReader(lr.r)
		if lr.err != nil {
			lr.err = fmt.Errorf("failed to create gzip reader: %w", lr.err)
			return 0, lr.err
		}
	}
	return lr.gr.Read(p)
}

func (lr *lazyReader) Close() error {
	if lr.gr != nil {
		if err := lr.gr.Close(); err != nil {
			return fmt.Errorf("failed to close gzip reader: %w", err)
		}
	}
	return nil
}
