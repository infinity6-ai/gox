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
