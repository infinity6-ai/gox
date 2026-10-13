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

func TestUnitReaderHeadTail(t *testing.T) {
	type testScenario struct {
		input          string
		headMaxSize    int
		tailMaxSize    int
		readBufSize    int
		expectedHead   string
		expectedTail   string
		expectedCount  uint64
		expectedStream string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		rht := ioz.NewReaderHeadTail(strings.NewReader(s.input), s.headMaxSize, s.tailMaxSize)

		bufSize := s.readBufSize
		if bufSize <= 0 {
			bufSize = 64
		}
		buf := make([]byte, bufSize)
		var readBytes []byte
		for {
			n, err := rht.Read(buf)
			if n > 0 {
				readBytes = append(readBytes, buf[:n]...)
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				require.NoError(t, err)
			}
		}

		require.Equal(t, s.expectedStream, string(readBytes))
		require.Equal(t, s.expectedCount, rht.Count())
		require.Equal(t, s.expectedHead, rht.HeadString())
		require.Equal(t, s.expectedTail, rht.TailString())
		require.Equal(t, []byte(s.expectedHead), rht.HeadBytes())
		require.Equal(t, []byte(s.expectedTail), rht.TailBytes())
	}

	t.Run("input smaller than head max size leaves tail empty", func(t *testing.T) {
		check(t, testScenario{
			input:          "abc",
			headMaxSize:    5,
			tailMaxSize:    5,
			readBufSize:    32,
			expectedHead:   "abc",
			expectedTail:   "",
			expectedCount:  3,
			expectedStream: "abc",
		})
	})

	t.Run("input equal to head max size leaves tail empty", func(t *testing.T) {
		check(t, testScenario{
			input:          "abcde",
			headMaxSize:    5,
			tailMaxSize:    5,
			readBufSize:    32,
			expectedHead:   "abcde",
			expectedTail:   "",
			expectedCount:  5,
			expectedStream: "abcde",
		})
	})

	t.Run("input between head and head plus tail populates partial tail with no overlay", func(t *testing.T) {
		check(t, testScenario{
			input:          "abcdefgh",
			headMaxSize:    5,
			tailMaxSize:    5,
			readBufSize:    32,
			expectedHead:   "abcde",
			expectedTail:   "fgh",
			expectedCount:  8,
			expectedStream: "abcdefgh",
		})
	})

	t.Run("input equal to head plus tail fills both completely", func(t *testing.T) {
		check(t, testScenario{
			input:          "0123456789",
			headMaxSize:    5,
			tailMaxSize:    5,
			readBufSize:    32,
			expectedHead:   "01234",
			expectedTail:   "56789",
			expectedCount:  10,
			expectedStream: "0123456789",
		})
	})

	t.Run("input larger than head plus tail slides tail to keep the last bytes", func(t *testing.T) {
		check(t, testScenario{
			input:          "0123456789abcdef",
			headMaxSize:    5,
			tailMaxSize:    5,
			readBufSize:    32,
			expectedHead:   "01234",
			expectedTail:   "bcdef",
			expectedCount:  16,
			expectedStream: "0123456789abcdef",
		})
	})

	t.Run("chunked reading with small 2-byte buffers matches single read", func(t *testing.T) {
		check(t, testScenario{
			input:          "0123456789abcdef",
			headMaxSize:    5,
			tailMaxSize:    5,
			readBufSize:    2,
			expectedHead:   "01234",
			expectedTail:   "bcdef",
			expectedCount:  16,
			expectedStream: "0123456789abcdef",
		})
	})

	t.Run("chunked reading with 1-byte buffer works accurately", func(t *testing.T) {
		check(t, testScenario{
			input:          "abcdefghijkl",
			headMaxSize:    4,
			tailMaxSize:    3,
			readBufSize:    1,
			expectedHead:   "abcd",
			expectedTail:   "jkl",
			expectedCount:  12,
			expectedStream: "abcdefghijkl",
		})
	})

	t.Run("head max size zero captures only tail", func(t *testing.T) {
		check(t, testScenario{
			input:          "abcdefghij",
			headMaxSize:    0,
			tailMaxSize:    4,
			readBufSize:    32,
			expectedHead:   "",
			expectedTail:   "ghij",
			expectedCount:  10,
			expectedStream: "abcdefghij",
		})
	})

	t.Run("tail max size zero captures only head", func(t *testing.T) {
		check(t, testScenario{
			input:          "abcdefghij",
			headMaxSize:    4,
			tailMaxSize:    0,
			readBufSize:    32,
			expectedHead:   "abcd",
			expectedTail:   "",
			expectedCount:  10,
			expectedStream: "abcdefghij",
		})
	})

	t.Run("negative sizes treated as zero", func(t *testing.T) {
		check(t, testScenario{
			input:          "abcdefghij",
			headMaxSize:    -1,
			tailMaxSize:    -2,
			readBufSize:    32,
			expectedHead:   "",
			expectedTail:   "",
			expectedCount:  10,
			expectedStream: "abcdefghij",
		})
	})
}

func TestUnitReaderHeadTailEdgeCases(t *testing.T) {
	t.Run("nil reader returns EOF and 0 count", func(t *testing.T) {
		rht := &ioz.ReaderHeadTail{HeadMaxSize: 5, TailMaxSize: 5}
		buf := make([]byte, 10)
		n, err := rht.Read(buf)
		require.Equal(t, 0, n)
		require.Equal(t, io.EOF, err)
		require.Equal(t, uint64(0), rht.Count())
	})
}

func TestUnitWriterHeadTail(t *testing.T) {
	type testScenario struct {
		chunks        []string
		headMaxSize   int
		tailMaxSize   int
		expectedHead  string
		expectedTail  string
		expectedCount uint64
		expectedTotal string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		var out bytes.Buffer
		wht := ioz.NewWriterHeadTail(&out, s.headMaxSize, s.tailMaxSize)

		for _, chunk := range s.chunks {
			n, err := wht.Write([]byte(chunk))
			require.NoError(t, err)
			require.Equal(t, len(chunk), n)
		}

		require.Equal(t, s.expectedTotal, out.String())
		require.Equal(t, s.expectedCount, wht.Count())
		require.Equal(t, s.expectedHead, wht.HeadString())
		require.Equal(t, s.expectedTail, wht.TailString())
		require.Equal(t, []byte(s.expectedHead), wht.HeadBytes())
		require.Equal(t, []byte(s.expectedTail), wht.TailBytes())
	}

	t.Run("single write smaller than head leaves tail empty", func(t *testing.T) {
		check(t, testScenario{
			chunks:        []string{"abc"},
			headMaxSize:   5,
			tailMaxSize:   5,
			expectedHead:  "abc",
			expectedTail:  "",
			expectedCount: 3,
			expectedTotal: "abc",
		})
	})

	t.Run("single write between head and head plus tail fills partial tail", func(t *testing.T) {
		check(t, testScenario{
			chunks:        []string{"abcdefg"},
			headMaxSize:   5,
			tailMaxSize:   5,
			expectedHead:  "abcde",
			expectedTail:  "fg",
			expectedCount: 7,
			expectedTotal: "abcdefg",
		})
	})

	t.Run("single write larger than head plus tail slides tail", func(t *testing.T) {
		check(t, testScenario{
			chunks:        []string{"0123456789abcdef"},
			headMaxSize:   5,
			tailMaxSize:   5,
			expectedHead:  "01234",
			expectedTail:  "bcdef",
			expectedCount: 16,
			expectedTotal: "0123456789abcdef",
		})
	})

	t.Run("multiple incremental writes straddle boundaries accurately", func(t *testing.T) {
		check(t, testScenario{
			chunks:        []string{"ab", "cde", "f", "ghi", "jklm"},
			headMaxSize:   5,
			tailMaxSize:   4,
			expectedHead:  "abcde",
			expectedTail:  "jklm",
			expectedCount: 13,
			expectedTotal: "abcdefghijklm",
		})
	})

	t.Run("head size zero captures only tail", func(t *testing.T) {
		check(t, testScenario{
			chunks:        []string{"abcdefghij"},
			headMaxSize:   0,
			tailMaxSize:   3,
			expectedHead:  "",
			expectedTail:  "hij",
			expectedCount: 10,
			expectedTotal: "abcdefghij",
		})
	})

	t.Run("tail size zero captures only head", func(t *testing.T) {
		check(t, testScenario{
			chunks:        []string{"abcdefghij"},
			headMaxSize:   4,
			tailMaxSize:   0,
			expectedHead:  "abcd",
			expectedTail:  "",
			expectedCount: 10,
			expectedTotal: "abcdefghij",
		})
	})
}

func TestUnitWriterHeadTailEdgeCases(t *testing.T) {
	t.Run("nil destination writer captures head and tail without panic", func(t *testing.T) {
		wht := ioz.NewWriterHeadTail(nil, 5, 5)
		n, err := wht.Write([]byte("0123456789abcdef"))
		require.NoError(t, err)
		require.Equal(t, 16, n)
		require.Equal(t, uint64(16), wht.Count())
		require.Equal(t, "01234", wht.HeadString())
		require.Equal(t, "bcdef", wht.TailString())
	})

}
