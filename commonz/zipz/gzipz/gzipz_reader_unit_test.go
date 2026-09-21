package gzipz_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/infinity6-ai/gox/commonz/zipz/gzipz"
	"github.com/stretchr/testify/require"
)

func TestUnitReaderValid(t *testing.T) {
	original := "Hello, this is a test of the lazy gzip reader!"
	compressed, err := gzipz.Gzip(original)
	require.NoError(t, err)

	buf := bytes.NewReader(compressed)
	reader := gzipz.Reader(buf)
	defer func() {
		err := reader.Close()
		require.NoError(t, err)
	}()

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, original, string(decompressed))
}

func TestUnitReaderInvalidHeader(t *testing.T) {
	// Provide invalid gzip header data
	invalidData := []byte("not a valid gzip stream")
	buf := bytes.NewReader(invalidData)

	// Instantiation should succeed without reading and without error
	reader := gzipz.Reader(buf)
	defer func() {
		err := reader.Close()
		require.NoError(t, err)
	}()

	// Error should be returned upon the first Read (via io.ReadAll)
	_, err := io.ReadAll(reader)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create gzip reader")
}

func TestUnitReaderCloseWithoutRead(t *testing.T) {
	buf := bytes.NewReader([]byte("some data"))
	reader := gzipz.Reader(buf)

	// Closing before any Read should be safe and return nil
	err := reader.Close()
	require.NoError(t, err)
}

func TestUnitReaderCloseAfterRead(t *testing.T) {
	original := "close after read test"
	compressed, err := gzipz.Gzip(original)
	require.NoError(t, err)

	buf := bytes.NewReader(compressed)
	reader := gzipz.Reader(buf)

	// Read some bytes first
	p := make([]byte, 5)
	n, err := reader.Read(p)
	require.NoError(t, err)
	require.Equal(t, 5, n)

	// Now close the reader, should succeed
	err = reader.Close()
	require.NoError(t, err)
}
