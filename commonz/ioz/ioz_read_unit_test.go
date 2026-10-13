package ioz_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/ioz"
	"github.com/stretchr/testify/require"
)

type singleByteReader struct {
	data string
	pos  int
}

func (s *singleByteReader) Read(p []byte) (int, error) {
	if s.pos >= len(s.data) {
		return 0, io.EOF
	}
	if len(p) == 0 {
		return 0, nil
	}
	p[0] = s.data[s.pos]
	s.pos++
	return 1, nil
}

type errReader struct {
	data   string
	pos    int
	errPos int
	err    error
}

func (e *errReader) Read(p []byte) (int, error) {
	if e.pos >= e.errPos {
		return 0, e.err
	}
	n := copy(p, e.data[e.pos:e.errPos])
	e.pos += n
	return n, nil
}

func TestUnitReadFull(t *testing.T) {
	type testScenario struct {
		reader       io.Reader
		size         int
		expectedData string
		errMsg       string
		isEOF        bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got, err := ioz.ReadFull(s.reader, s.size)
		if s.errMsg != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.errMsg)
			if s.isEOF {
				require.True(t, errors.Is(err, io.EOF))
			}
			require.Equal(t, s.expectedData, string(got))
			return
		}
		require.NoError(t, err)
		require.Equal(t, s.expectedData, string(got))
	}

	t.Run("reads exact buffer size", func(t *testing.T) {
		check(t, testScenario{
			reader:       strings.NewReader("hello world"),
			size:         11,
			expectedData: "hello world",
		})
	})

	t.Run("reads partial stream up to requested size", func(t *testing.T) {
		check(t, testScenario{
			reader:       strings.NewReader("hello world extra data"),
			size:         5,
			expectedData: "hello",
		})
	})

	t.Run("returns partial data and EOF when stream is shorter than size", func(t *testing.T) {
		check(t, testScenario{
			reader:       strings.NewReader("short"),
			size:         10,
			expectedData: "short",
			errMsg:       "failed to read full buffer: EOF",
			isEOF:        true,
		})
	})

	t.Run("reads incrementally from chunked reader", func(t *testing.T) {
		check(t, testScenario{
			reader:       &singleByteReader{data: "singlebytechunks"},
			size:         10,
			expectedData: "singlebyte",
		})
	})

	t.Run("returns empty slice on size zero", func(t *testing.T) {
		check(t, testScenario{
			reader:       strings.NewReader("data"),
			size:         0,
			expectedData: "",
		})
	})

	t.Run("returns empty slice on negative size", func(t *testing.T) {
		check(t, testScenario{
			reader:       strings.NewReader("data"),
			size:         -5,
			expectedData: "",
		})
	})

	t.Run("returns error when reader is nil", func(t *testing.T) {
		check(t, testScenario{
			reader: nil,
			size:   10,
			errMsg: "reader cannot be nil",
		})
	})

	t.Run("returns partial data and error when reader fails", func(t *testing.T) {
		check(t, testScenario{
			reader: &errReader{
				data:   "prefix data",
				errPos: 6,
				err:    errors.New("i/o failure"),
			},
			size:         10,
			expectedData: "prefix",
			errMsg:       "failed to read full buffer: i/o failure",
		})
	})
}
