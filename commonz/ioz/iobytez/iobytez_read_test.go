package iobytez_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	i "github.com/infinity6-ai/gox/commonz/ioz/iobytez"
)

func TestUnitRead_ReadAll(t *testing.T) {
	ctx := context.Background()

	opts := &i.Options{}
	data := "hello world"
	r := strings.NewReader(data)
	n, err := i.Read(ctx, r, opts)
	require.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, data, string(opts.Out))
}

func TestUnitRead_WithInitialData(t *testing.T) {
	ctx := context.Background()

	initial := "preexisting "
	opts := &i.Options{
		Out: []byte(initial),
	}
	data := "hello world"
	r := strings.NewReader(data)
	n, err := i.Read(ctx, r, opts)
	require.NoError(t, err)
	assert.Equal(t, len(data), n)
	expected := initial + data
	assert.Equal(t, expected, string(opts.Out))
}

func TestUnitRead_WithMin(t *testing.T) {
	ctx := context.Background()

	opts := &i.Options{Min: 5}
	data := "hello"
	r := strings.NewReader(data)
	n, err := i.Read(ctx, r, opts)
	require.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, data, string(opts.Out))
}

func TestUnitRead_MinNotMet(t *testing.T) {
	ctx := context.Background()

	opts := &i.Options{Min: 10}
	data := "hello"
	r := strings.NewReader(data)
	_, err := i.Read(ctx, r, opts)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
}

func TestUnitRead_WithMax(t *testing.T) {
	ctx := context.Background()

	opts := &i.Options{Max: 5}
	data := "hello world"
	r := strings.NewReader(data)
	n, err := i.Read(ctx, r, opts)
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "hello", string(opts.Out))
}

func TestUnitRead_WithMinAndMax(t *testing.T) {
	ctx := context.Background()

	opts := &i.Options{Min: 5, Max: 10}
	data := "hello world this is too long"
	r := strings.NewReader(data)
	n, err := i.Read(ctx, r, opts)
	require.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, "hello worl", string(opts.Out))
}

func TestUnitRead_EmptyReaderReturnsEOF(t *testing.T) {
	ctx := context.Background()

	opts := &i.Options{}
	data := ""
	r := strings.NewReader(data)
	n, err := i.Read(ctx, r, opts)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)
	assert.Empty(t, opts.Out)
}

func TestUnitRead_ReadInChunksUntilEOF(t *testing.T) {
	ctx := context.Background()

	fullData := "abcdefghijklmnopqrstuvwxyz"
	r := strings.NewReader(fullData)
	var totalReadBytes []byte

	// First chunk
	opts1 := &i.Options{Max: 10}
	n1, err := i.Read(ctx, r, opts1)
	require.NoError(t, err, "unexpected error on first read")
	assert.Equal(t, 10, n1, "expected to read 10 bytes on first chunk")
	expected1 := "abcdefghij"
	assert.Equal(t, expected1, string(opts1.Out), "expected first chunk")
	totalReadBytes = append(totalReadBytes, opts1.Out...)

	// Second chunk
	opts2 := &i.Options{Max: 10}
	n2, err := i.Read(ctx, r, opts2)
	require.NoError(t, err, "unexpected error on second read")
	assert.Equal(t, 10, n2, "expected to read 10 bytes on second chunk")
	expected2 := "klmnopqrst"
	assert.Equal(t, expected2, string(opts2.Out), "expected second chunk")
	totalReadBytes = append(totalReadBytes, opts2.Out...)

	// Third chunk (remaining data)
	opts3 := &i.Options{Max: 10}
	n3, err := i.Read(ctx, r, opts3)
	require.NoError(t, err, "unexpected error on third read")
	assert.Equal(t, 6, n3, "expected to read 6 bytes on third chunk")
	expected3 := "uvwxyz"
	assert.Equal(t, expected3, string(opts3.Out), "expected third chunk")
	totalReadBytes = append(totalReadBytes, opts3.Out...)

	// Fourth read (should be EOF)
	opts4 := &i.Options{Max: 10}
	n4, err := i.Read(ctx, r, opts4)
	assert.Equal(t, io.EOF, err, "expected io.EOF on fourth read")
	assert.Equal(t, 0, n4, "expected to read 0 bytes on fourth read")
	assert.Empty(t, opts4.Out, "expected empty buffer on fourth read")

	// Verify total data read
	assert.Equal(t, fullData, string(totalReadBytes), "expected total data")
}

func TestUnitRead_EmptyReaderWithMin(t *testing.T) {
	ctx := context.Background()

	opts := &i.Options{Min: 1}
	data := ""
	r := strings.NewReader(data)
	_, err := i.Read(ctx, r, opts)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
}

// mockReader is a test helper that wraps an io.Reader and counts the total bytes read.
type mockReader struct {
	io.Reader
	readCount int
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	n, err = m.Reader.Read(p)
	m.readCount += n
	return
}

func TestUnitRead_OriginalReaderNotOverRead(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		inputData   string
		options     *i.Options
		expectedN   int
		expectedOut string
		expectedErr error
		expectedReadCount int // How many bytes the mockReader should have read
	}{
		{
			name:        "Read all with no specific options",
			inputData:   "hello world",
			options:     &i.Options{},
			expectedN:   11,
			expectedOut: "hello world",
			expectedErr: nil,
			expectedReadCount: 11,
		},
		{
			name:        "Read with Max limit",
			inputData:   "hello world",
			options:     &i.Options{Max: 5},
			expectedN:   5,
			expectedOut: "hello",
			expectedErr: nil,
			expectedReadCount: 5, // Should only read up to Max
		},
		{
			name:        "Read with Min and Max (exact)",
			inputData:   "1234567890",
			options:     &i.Options{Min: 5, Max: 10},
			expectedN:   10,
			expectedOut: "1234567890",
			expectedErr: nil,
			expectedReadCount: 10,
		},
		{
			name:        "Read with Min and Max (over Max, truncate)",
			inputData:   "1234567890abcdef",
			options:     &i.Options{Min: 5, Max: 10},
			expectedN:   10,
			expectedOut: "1234567890",
			expectedErr: nil,
			expectedReadCount: 10, // Should read only up to Max
		},
		{
			name:        "Min not met",
			inputData:   "short",
			options:     &i.Options{Min: 10},
			expectedN:   5,
			expectedOut: "short",
			expectedErr: io.ErrUnexpectedEOF,
			expectedReadCount: 5, // Reads all available, but min not met
		},
		{
			name:        "Empty reader returns EOF",
			inputData:   "",
			options:     &i.Options{},
			expectedN:   0,
			expectedOut: "",
			expectedErr: io.EOF,
			expectedReadCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockReader{Reader: strings.NewReader(tc.inputData)}
			n, err := i.Read(ctx, mock, tc.options)

			assert.Equal(t, tc.expectedN, n, "mismatched bytes read by i.Read")
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err, "mismatched error from i.Read")
			} else {
				assert.NoError(t, err, "unexpected error from i.Read")
			}
			assert.Equal(t, tc.expectedOut, string(tc.options.Out), "mismatched output data")
			assert.Equal(t, tc.expectedReadCount, mock.readCount, "original reader over-read")
		})
	}
}
