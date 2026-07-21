package iorunez_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"golang.org/x/text/encoding/unicode"

	i "github.com/infinity6-ai/gox/commonz/ioz/iorunez"
)

func TestUnitRead_ReadAll(t *testing.T) {
	ctx := context.Background()

	opts := &i.Options{}
	data := "hello world"
	r := strings.NewReader(data)
	n, err := i.Read(ctx, r, opts)
	require.NoError(t, err)
	assert.Equal(t, len([]rune(data)), n)
	assert.Equal(t, data, string(opts.Out))
}

func TestUnitRead_WithInitialData(t *testing.T) {
	ctx := context.Background()

	initial := "preexisting "
	opts := &i.Options{
		Out: []rune(initial),
	}
	data := "hello world"
	r := strings.NewReader(data)
	n, err := i.Read(ctx, r, opts)
	require.NoError(t, err)
	assert.Equal(t, len([]rune(data)), n)
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
	assert.Equal(t, len([]rune(data)), n)
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
	var totalReadRunes []rune

	opts := &i.Options{}

	// First chunk
	opts.Max = 10
	opts.Out = nil
	n1, err := i.Read(ctx, r, opts)
	require.NoError(t, err, "unexpected error on first read")
	assert.Equal(t, 10, n1, "expected to read 10 runes on first chunk")
	expected1 := "abcdefghij"
	assert.Equal(t, expected1, string(opts.Out), "expected first chunk")
	totalReadRunes = append(totalReadRunes, opts.Out...)

	// Second chunk
	opts.Max = 10
	opts.Out = nil
	n2, err := i.Read(ctx, r, opts)
	require.NoError(t, err, "unexpected error on second read")
	assert.Equal(t, 10, n2, "expected to read 10 runes on second chunk")
	expected2 := "klmnopqrst"
	assert.Equal(t, expected2, string(opts.Out), "expected second chunk")
	totalReadRunes = append(totalReadRunes, opts.Out...)

	// Third chunk (remaining data)
	opts.Max = 10
	opts.Out = nil
	n3, err := i.Read(ctx, r, opts)
	require.NoError(t, err, "unexpected error on third read")
	assert.Equal(t, 6, n3, "expected to read 6 runes on third chunk")
	expected3 := "uvwxyz"
	assert.Equal(t, expected3, string(opts.Out), "expected third chunk")
	totalReadRunes = append(totalReadRunes, opts.Out...)

	// Fourth read (should be EOF)
	opts.Max = 10
	opts.Out = nil
	n4, err := i.Read(ctx, r, opts)
	assert.Equal(t, io.EOF, err, "expected io.EOF on fourth read")
	assert.Equal(t, 0, n4, "expected to read 0 runes on fourth read")
	assert.Empty(t, opts.Out, "expected empty buffer on fourth read")

	// Verify total data read
	assert.Equal(t, fullData, string(totalReadRunes), "expected total data")
}

func TestUnitRead_EmptyReaderWithMin(t *testing.T) {
	ctx := context.Background()

	opts := &i.Options{Min: 1}
	data := ""
	r := strings.NewReader(data)
	_, err := i.Read(ctx, r, opts)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
}

func TestUnitRead_WithUTF16BE(t *testing.T) {
	ctx := context.Background()

	utf16be := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	encoder := utf16be.NewEncoder()
	encoded, err := encoder.String("hello world")
	require.NoError(t, err)

	opts := &i.Options{
		Encoding: utf16be,
	}
	r := strings.NewReader(encoded)
	n, err := i.Read(ctx, r, opts)
	require.NoError(t, err)
	assert.Equal(t, 11, n)
	assert.Equal(t, "hello world", string(opts.Out))
}

func TestUnitRead_WithUTF16LE(t *testing.T) {
	ctx := context.Background()

	utf16le := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	encoder := utf16le.NewEncoder()
	encoded, err := encoder.String("hello world")
	require.NoError(t, err)

	opts := &i.Options{
		Encoding: utf16le,
	}
	r := strings.NewReader(encoded)
	n, err := i.Read(ctx, r, opts)
	require.NoError(t, err)
	assert.Equal(t, 11, n)
	assert.Equal(t, "hello world", string(opts.Out))
}
