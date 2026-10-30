package gzipzsmartwriter_test

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/zipz/gzipz/gzipzsmartwriter"
	"github.com/stretchr/testify/require"
)

type trackCloserWriter struct {
	bytes.Buffer
	closed bool
}

func (t *trackCloserWriter) Close() error {
	t.closed = true
	return nil
}

type errorWriter struct {
	failOnWrite bool
}

func (e *errorWriter) Write(p []byte) (int, error) {
	if e.failOnWrite {
		return 0, fmt.Errorf("write error")
	}
	return len(p), nil
}

type trackFlushWriter struct {
	bytes.Buffer
	flushed bool
}

func (t *trackFlushWriter) Flush() error {
	t.flushed = true
	return nil
}

func TestUnitGzipzSmartWriterDecisionAndCompression(t *testing.T) {
	type testScenario struct {
		name         string
		input        []byte
		chunkSize    int
		head         int
		wantCompress bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		underlying := &trackCloserWriter{}
		var decisionCalled bool
		var decisionValue bool
		var underlyingBytesAtDecision int

		callback := func(compress bool) {
			decisionCalled = true
			decisionValue = compress
			underlyingBytesAtDecision = underlying.Len()
		}

		opts := gzipzsmartwriter.Opts{
			Head:     s.head,
			Decision: callback,
		}

		w := gzipzsmartwriter.New(underlying, opts)
		require.NotNil(t, w)

		// Write in chunks
		chunkSize := s.chunkSize
		if chunkSize <= 0 {
			chunkSize = len(s.input)
		}
		if chunkSize == 0 && len(s.input) == 0 {
			// empty write test
		} else {
			for offset := 0; offset < len(s.input); offset += chunkSize {
				end := offset + chunkSize
				if end > len(s.input) {
					end = len(s.input)
				}
				n, err := w.Write(s.input[offset:end])
				require.NoError(t, err)
				require.Equal(t, end-offset, n)
			}
		}

		err := w.Close()
		require.NoError(t, err)

		// Underlying writer must NEVER be closed by smartGzipWriter.Close()
		require.False(t, underlying.closed, "underlying writer must not be closed")

		// Decision callback must have been called
		require.True(t, decisionCalled, "decision callback must be called")
		require.Equal(t, s.wantCompress, decisionValue, "decision value mismatch")

		// Decision callback must be invoked before ANY byte is written to the underlying writer
		require.Equal(t, 0, underlyingBytesAtDecision, "no bytes should have been written to underlying writer before decision")

		// Verify content integrity
		if s.wantCompress {
			// Should be valid gzip stream
			gzReader, err := gzip.NewReader(bytes.NewReader(underlying.Bytes()))
			require.NoError(t, err)
			decompressed, err := io.ReadAll(gzReader)
			require.NoError(t, err)
			require.NoError(t, gzReader.Close())
			require.Equal(t, s.input, decompressed)
			require.Less(t, underlying.Len(), len(s.input), "compressed output should be smaller than input")
		} else {
			// Should be identical raw bytes
			if len(s.input) == 0 {
				require.Empty(t, underlying.Bytes())
			} else {
				require.Equal(t, s.input, underlying.Bytes())
			}
		}
	}

	// 1. Compressible large text (> 1024 bytes) written in one chunk
	t.Run("Compressible text single write larger than Head", func(t *testing.T) {
		input := []byte(strings.Repeat("Smart gzip compression saves space on repetitive text! ", 50))
		check(t, testScenario{
			name:         "Compressible text single write larger than Head",
			input:        input,
			chunkSize:    len(input),
			wantCompress: true,
		})
	})

	// 2. Compressible large text written in small chunks (100 bytes each)
	t.Run("Compressible text small chunks larger than Head", func(t *testing.T) {
		input := []byte(strings.Repeat("Smart gzip compression saves space on repetitive text! ", 50))
		check(t, testScenario{
			name:         "Compressible text small chunks larger than Head",
			input:        input,
			chunkSize:    100,
			wantCompress: true,
		})
	})

	// 3. Incompressible random data (> 1024 bytes)
	t.Run("Incompressible random data larger than Head", func(t *testing.T) {
		randomData := make([]byte, 2048)
		_, err := rand.Read(randomData)
		require.NoError(t, err)
		check(t, testScenario{
			name:         "Incompressible random data larger than Head",
			input:        randomData,
			chunkSize:    len(randomData),
			wantCompress: false,
		})
	})

	// 4. Incompressible random data written in small chunks
	t.Run("Incompressible random data small chunks larger than Head", func(t *testing.T) {
		randomData := make([]byte, 2048)
		_, err := rand.Read(randomData)
		require.NoError(t, err)
		check(t, testScenario{
			name:         "Incompressible random data small chunks larger than Head",
			input:        randomData,
			chunkSize:    128,
			wantCompress: false,
		})
	})

	// 5. Compressible data smaller than Head (< 1024 bytes)
	t.Run("Compressible text smaller than Head", func(t *testing.T) {
		input := []byte(strings.Repeat("Compress me please! ", 20)) // ~400 bytes, highly compressible
		check(t, testScenario{
			name:         "Compressible text smaller than Head",
			input:        input,
			chunkSize:    len(input),
			wantCompress: true,
		})
	})

	// 6. Incompressible data smaller than Head (< 1024 bytes)
	t.Run("Small incompressible string smaller than Head", func(t *testing.T) {
		input := []byte("short text") // 10 bytes: gzip overhead exceeds original size
		check(t, testScenario{
			name:         "Small incompressible string smaller than Head",
			input:        input,
			chunkSize:    len(input),
			wantCompress: false,
		})
	})

	// 7. Empty data
	t.Run("Empty stream", func(t *testing.T) {
		check(t, testScenario{
			name:         "Empty stream",
			input:        []byte{},
			chunkSize:    0,
			wantCompress: false,
		})
	})

	// 8. Custom Head size (2048 bytes)
	t.Run("Custom larger Head size", func(t *testing.T) {
		input := []byte(strings.Repeat("Custom head size test repeating pattern. ", 100))
		check(t, testScenario{
			name:         "Custom larger Head size",
			input:        input,
			head:         2048,
			chunkSize:    256,
			wantCompress: true,
		})
	})
}

func TestUnitGzipzSmartWriterErrorsAndEdgeCases(t *testing.T) {
	t.Run("Write on closed writer returns error", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := gzipzsmartwriter.New(buf, gzipzsmartwriter.Opts{})
		err := w.Close()
		require.NoError(t, err)

		n, err := w.Write([]byte("data after close"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "closed")
		require.Equal(t, 0, n)
	})

	t.Run("Multiple Close calls are idempotent", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := gzipzsmartwriter.New(buf, gzipzsmartwriter.Opts{})
		_, err := w.Write([]byte("some data"))
		require.NoError(t, err)

		require.NoError(t, w.Close())
		require.NoError(t, w.Close())
	})

	t.Run("Write error on underlying writer propagates", func(t *testing.T) {
		ew := &errorWriter{failOnWrite: true}
		w := gzipzsmartwriter.New(ew, gzipzsmartwriter.Opts{
			Head: 1024,
		})

		// Write incompressible data larger than Head to trigger underlying writer error
		randomData := make([]byte, 2048)
		_, _ = rand.Read(randomData)
		_, err := w.Write(randomData)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to write")
	})

	t.Run("Flush before decision forces decision and flushes", func(t *testing.T) {
		underlying := &trackFlushWriter{}
		var decisionCalled bool
		var decisionValue bool

		w := gzipzsmartwriter.New(underlying, gzipzsmartwriter.Opts{
			Head: 2048,
			Decision: func(compress bool) {
				decisionCalled = true
				decisionValue = compress
			},
		})

		flusher, ok := w.(interface{ Flush() error })
		require.True(t, ok)

		// Write compressible data less than Head
		input := []byte(strings.Repeat("flush this data early! ", 20))
		_, err := w.Write(input)
		require.NoError(t, err)

		require.False(t, decisionCalled)

		err = flusher.Flush()
		require.NoError(t, err)
		require.True(t, decisionCalled)
		require.True(t, decisionValue)
		require.True(t, underlying.flushed)

		err = w.Close()
		require.NoError(t, err)

		// Verify decompressed output matches
		gzReader, err := gzip.NewReader(bytes.NewReader(underlying.Bytes()))
		require.NoError(t, err)
		decompressed, err := io.ReadAll(gzReader)
		require.NoError(t, err)
		require.NoError(t, gzReader.Close())
		require.Equal(t, input, decompressed)
	})

	t.Run("Flush on closed writer returns error", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := gzipzsmartwriter.New(buf, gzipzsmartwriter.Opts{})
		require.NoError(t, w.Close())

		flusher, ok := w.(interface{ Flush() error })
		require.True(t, ok)
		err := flusher.Flush()
		require.Error(t, err)
		require.Contains(t, err.Error(), "closed")
	})

	t.Run("Reset allows writer reuse", func(t *testing.T) {
		buf1 := &bytes.Buffer{}
		w := gzipzsmartwriter.New(buf1, gzipzsmartwriter.Opts{})
		data1 := []byte(strings.Repeat("first writer data! ", 50))
		_, err := w.Write(data1)
		require.NoError(t, err)
		require.NoError(t, w.Close())

		resettable, ok := w.(interface{ Reset(io.Writer) })
		require.True(t, ok)

		buf2 := &bytes.Buffer{}
		resettable.Reset(buf2)

		data2 := []byte(strings.Repeat("second writer data! ", 50))
		_, err = w.Write(data2)
		require.NoError(t, err)
		require.NoError(t, w.Close())

		// Verify second stream is valid gzip
		gzReader, err := gzip.NewReader(bytes.NewReader(buf2.Bytes()))
		require.NoError(t, err)
		decompressed, err := io.ReadAll(gzReader)
		require.NoError(t, err)
		require.NoError(t, gzReader.Close())
		require.Equal(t, data2, decompressed)
	})
}
