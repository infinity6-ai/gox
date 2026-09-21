package gzipz_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/zipz/gzipz"
	"github.com/stretchr/testify/require"
)

func TestUnitGzipGunzipBasic(t *testing.T) {
	type testScenario struct {
		name  string
		input string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		// Test Gzip compression
		compressed, err := gzipz.Gzip(s.input)
		require.NoError(t, err)
		require.NotEmpty(t, compressed)

		// Test Gunzip decompression
		decompressed, err := gzipz.Gunzip(compressed)
		require.NoError(t, err)
		require.Equal(t, s.input, decompressed.String())
	}

	t.Run("Standard text string", func(t *testing.T) {
		check(t, testScenario{
			name:  "Standard text string",
			input: "Hello, this is a standard string to compress!",
		})
	})

	t.Run("Empty string", func(t *testing.T) {
		check(t, testScenario{
			name:  "Empty string",
			input: "",
		})
	})

	t.Run("Long repeating string", func(t *testing.T) {
		check(t, testScenario{
			name:  "Long repeating string",
			input: strings.Repeat("A very long repeating string. ", 100),
		})
	})
}

func TestUnitGzipByteSlice(t *testing.T) {
	original := []byte("some byte data containing special chars: \x00\x01\x02")
	compressed, err := gzipz.Gzip(original)
	require.NoError(t, err)
	require.NotEmpty(t, compressed)

	decompressed, err := gzipz.Gunzip(compressed)
	require.NoError(t, err)
	require.Equal(t, original, decompressed.Bytes())
}

func TestUnitMustGzipMustGunzip(t *testing.T) {
	input := "must gzip and must gunzip test data"
	compressed := gzipz.MustGzip(input)
	require.NotEmpty(t, compressed)

	decompressed := gzipz.MustGunzip(compressed)
	require.Equal(t, input, decompressed.String())
}

func TestUnitMustGunzipPanicOnInvalidData(t *testing.T) {
	invalidGzipData := []byte("not a valid gzip stream")
	require.Panics(t, func() {
		gzipz.MustGunzip(invalidGzipData)
	})
}

func TestUnitGunzipInvalidDataReturnsError(t *testing.T) {
	invalidGzipData := []byte("not a valid gzip stream")
	_, err := gzipz.Gunzip(invalidGzipData)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to gunzip")
}

func TestUnitGunzipLimitedPanicOnExceedingSize(t *testing.T) {
	input := "some reasonably long content to test size limitations in gunzip"
	compressed, err := gzipz.Gzip(input)
	require.NoError(t, err)

	// Set maxSize less than the uncompressed content length to trigger panic.
	maxSize := 10
	require.Panics(t, func() {
		_, _ = gzipz.GunzipLimited(compressed, maxSize)
	})
}

func TestUnitMustReader(t *testing.T) {
	type testScenario struct {
		name        string
		inputData   string
		isValidGzip bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.isValidGzip {
			compressed := gzipz.MustGzip(s.inputData)
			buf := bytes.NewReader(compressed)

			reader := gzipz.MustReader(buf)
			require.NotNil(t, reader)
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			require.NoError(t, err)
			require.Equal(t, s.inputData, string(decompressed))
		} else {
			invalidBuf := bytes.NewReader([]byte(s.inputData))
			require.Panics(t, func() {
				_ = gzipz.MustReader(invalidBuf)
			})
		}
	}

	t.Run("Valid gzip stream", func(t *testing.T) {
		check(t, testScenario{
			name:        "Valid gzip stream",
			inputData:   "data for must reader",
			isValidGzip: true,
		})
	})

	t.Run("Invalid reader triggers panic", func(t *testing.T) {
		check(t, testScenario{
			name:        "Invalid reader triggers panic",
			inputData:   "not a gzip stream",
			isValidGzip: false,
		})
	})
}
