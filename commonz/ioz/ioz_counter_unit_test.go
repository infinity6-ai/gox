package ioz_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/ioz"
	"github.com/stretchr/testify/require"
)

type counterErrWriter struct {
	err error
}

func (e *counterErrWriter) Write(p []byte) (int, error) {
	return 0, e.err
}

func TestUnitReaderCounter(t *testing.T) {
	type testScenario struct {
		reader        io.Reader
		chunkSize     int
		expectedData  string
		expectedCount uint64
		errMsg        string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		rc := ioz.NewReaderCounter(s.reader)
		require.NotNil(t, rc)

		chunkSize := s.chunkSize
		if chunkSize <= 0 {
			chunkSize = 32
		}
		buf := make([]byte, chunkSize)
		var readBytes []byte
		for {
			n, err := rc.Read(buf)
			if n > 0 {
				readBytes = append(readBytes, buf[:n]...)
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				if s.errMsg != "" {
					require.Error(t, err)
					require.Contains(t, err.Error(), s.errMsg)
					break
				}
				require.NoError(t, err)
			}
		}

		require.Equal(t, s.expectedData, string(readBytes))
		require.Equal(t, s.expectedCount, rc.Count())
	}

	t.Run("single read reads full content and updates count", func(t *testing.T) {
		check(t, testScenario{
			reader:        strings.NewReader("hello world"),
			chunkSize:     64,
			expectedData:  "hello world",
			expectedCount: 11,
		})
	})

	t.Run("chunked reads accumulate count accurately", func(t *testing.T) {
		check(t, testScenario{
			reader:        strings.NewReader("0123456789"),
			chunkSize:     2,
			expectedData:  "0123456789",
			expectedCount: 10,
		})
	})

	t.Run("reader error returns wrapped error and retains count", func(t *testing.T) {
		check(t, testScenario{
			reader: &errReader{
				data:   "abcdef",
				pos:    0,
				errPos: 3,
				err:    errors.New("read failed"),
			},
			chunkSize:     2,
			expectedData:  "abc",
			expectedCount: 3,
			errMsg:        "failed to read: read failed",
		})
	})

	t.Run("nil reader returns EOF and 0 count", func(t *testing.T) {
		rc := ioz.NewReaderCounter(nil)
		buf := make([]byte, 10)
		n, err := rc.Read(buf)
		require.Equal(t, 0, n)
		require.Equal(t, io.EOF, err)
		require.Equal(t, uint64(0), rc.Count())
	})

	t.Run("nil receiver handles Read and Count safely", func(t *testing.T) {
		var nilRC *ioz.ReaderCounter
		require.Equal(t, uint64(0), nilRC.Count())
		buf := make([]byte, 10)
		n, err := nilRC.Read(buf)
		require.Equal(t, 0, n)
		require.Equal(t, io.EOF, err)
	})
}

func TestUnitWriterCounter(t *testing.T) {
	type testScenario struct {
		chunks        []string
		expectedData  string
		expectedCount uint64
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		var out bytes.Buffer
		wc := ioz.NewWriterCounter(&out)
		require.NotNil(t, wc)

		for _, chunk := range s.chunks {
			n, err := wc.Write([]byte(chunk))
			require.NoError(t, err)
			require.Equal(t, len(chunk), n)
		}

		require.Equal(t, s.expectedData, out.String())
		require.Equal(t, s.expectedCount, wc.Count())
	}

	t.Run("single write writes data and updates count", func(t *testing.T) {
		check(t, testScenario{
			chunks:        []string{"hello world"},
			expectedData:  "hello world",
			expectedCount: 11,
		})
	})

	t.Run("multiple writes accumulate count accurately", func(t *testing.T) {
		check(t, testScenario{
			chunks:        []string{"foo", "bar", "baz"},
			expectedData:  "foobarbaz",
			expectedCount: 9,
		})
	})

	t.Run("nil writer counts bytes without error", func(t *testing.T) {
		wc := ioz.NewWriterCounter(nil)
		n, err := wc.Write([]byte("counting bytes"))
		require.NoError(t, err)
		require.Equal(t, 14, n)
		require.Equal(t, uint64(14), wc.Count())
	})

	t.Run("writer error returns wrapped error", func(t *testing.T) {
		wc := ioz.NewWriterCounter(&counterErrWriter{err: errors.New("write failure")})
		n, err := wc.Write([]byte("data"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to write: write failure")
		require.Equal(t, 0, n)
		require.Equal(t, uint64(0), wc.Count())
	})

	t.Run("nil receiver handles Write and Count safely", func(t *testing.T) {
		var nilWC *ioz.WriterCounter
		require.Equal(t, uint64(0), nilWC.Count())
		n, err := nilWC.Write([]byte("data"))
		require.NoError(t, err)
		require.Equal(t, 0, n)
	})
}
