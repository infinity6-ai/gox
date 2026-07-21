package iorunez_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/infinity6-ai/gox/commonz/ioz/iorunez"
)

func checkRead(t *testing.T, readerData string, opts *iorunez.Options, dataFirst string, errFirst error, dataSecond string, errSecond error) {
	t.Helper()
	ctx := t.Context()
	oldData := string(opts.Out)
	r := strings.NewReader(readerData)

	n, err := iorunez.Read(ctx, r, opts)
	require.ErrorIs(t, err, errFirst)
	assert.Equal(t, oldData+dataFirst, string(opts.Out))
	assert.Equal(t, len(dataFirst), n)

	// Keep track of the runes read in the first call
	firstReadCount := n
	oldSize := len(opts.Out)

	n, err = iorunez.Read(ctx, r, opts)
	require.ErrorIs(t, err, errSecond)

	// We need to check against what's new.
	if len(opts.Out) > oldSize {
		assert.Equal(t, dataSecond, string(opts.Out[oldSize:]))
		assert.Equal(t, len(dataSecond), n)
	} else {
		assert.Equal(t, "", dataSecond)
		assert.Equal(t, 0, n)
	}

	// Verify what's left in the reader
	// The reader is consumed by iorunez.Read, but we can calculate what should be left
	totalReadRunes := firstReadCount
	runes := []rune(readerData)
	if len(runes) > totalReadRunes {
		remainingData := string(runes[totalReadRunes:])
		remaning, err := io.ReadAll(r)
		require.NoError(t, err)
		assert.Equal(t, remainingData, string(remaning))
	}

	n, err = r.Read([]byte{0})
	require.ErrorIs(t, err, io.EOF)
	assert.Equal(t, 0, n)
}

func TestUnitRead_ReadAll(t *testing.T) {
	checkRead(t, "hello world", &iorunez.Options{}, "hello world", nil, "", io.EOF)
}

func TestUnitRead_WithInitialData(t *testing.T) {
	checkRead(t, "hello world", &iorunez.Options{Out: []rune("preexisting ")}, "hello world", nil, "", io.EOF)
}

func TestUnitRead_WithMin(t *testing.T) {
	ctx := context.Background()

	opts := &iorunez.Options{Min: 5}
	data := "hello"
	r := strings.NewReader(data)
	n, err := iorunez.Read(ctx, r, opts)
	require.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, data, string(opts.Out))
}

func TestUnitRead_MinNotMet(t *testing.T) {
	ctx := context.Background()

	opts := &iorunez.Options{Min: 10}
	data := "hello"
	r := strings.NewReader(data)
	_, err := iorunez.Read(ctx, r, opts)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
}

func TestUnitRead_WithMax(t *testing.T) {
	ctx := context.Background()

	opts := &iorunez.Options{Max: 5}
	data := "hello world"
	r := strings.NewReader(data)
	n, err := iorunez.Read(ctx, r, opts)
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "hello", string(opts.Out))
}

func TestUnitRead_WithMinAndMax(t *testing.T) {
	ctx := context.Background()

	opts := &iorunez.Options{Min: 5, Max: 10}
	data := "hello world this is too long"
	r := strings.NewReader(data)
	n, err := iorunez.Read(ctx, r, opts)
	require.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, "hello worl", string(opts.Out))
}

func TestUnitRead_Empty(t *testing.T) {
	checkRead(t, "", &iorunez.Options{}, "", io.EOF, "", io.EOF)
}

func TestUnitRead_ReadInChunksUntilEOF(t *testing.T) {
	ctx := context.Background()

	fullData := "abcdefghijklmnopqrstuvwxyz"
	r := strings.NewReader(fullData)
	var totalReadRunes []rune

	// First chunk
	opts1 := &iorunez.Options{Max: 10}
	n1, err := iorunez.Read(ctx, r, opts1)
	require.NoError(t, err, "unexpected error on first read")
	assert.Equal(t, 10, n1, "expected to read 10 runes on first chunk")
	expected1 := "abcdefghij"
	assert.Equal(t, expected1, string(opts1.Out), "expected first chunk")
	totalReadRunes = append(totalReadRunes, opts1.Out...)

	// The reader is consumed, so for the next chunk, we pass the same reader
	// but the state of the reader is advanced.

	// Second chunk
	opts2 := &iorunez.Options{Max: 10}
	n2, err := iorunez.Read(ctx, r, opts2)
	require.NoError(t, err, "unexpected error on second read")
	assert.Equal(t, 10, n2, "expected to read 10 runes on second chunk")
	expected2 := "klmnopqrst"
	assert.Equal(t, expected2, string(opts2.Out), "expected second chunk")
	totalReadRunes = append(totalReadRunes, opts2.Out...)

	// Third chunk (remaining data)
	opts3 := &iorunez.Options{Max: 10}
	n3, err := iorunez.Read(ctx, r, opts3)
	require.NoError(t, err, "unexpected error on third read")
	assert.Equal(t, 6, n3, "expected to read 6 runes on third chunk")
	expected3 := "uvwxyz"
	assert.Equal(t, expected3, string(opts3.Out), "expected third chunk")
	totalReadRunes = append(totalReadRunes, opts3.Out...)

	// Fourth read (should be EOF)
	opts4 := &iorunez.Options{Max: 10}
	n4, err := iorunez.Read(ctx, r, opts4)
	assert.Equal(t, io.EOF, err, "expected io.EOF on fourth read")
	assert.Equal(t, 0, n4, "expected to read 0 runes on fourth read")
	assert.Empty(t, opts4.Out, "expected empty buffer on fourth read")

	// Verify total data read
	assert.Equal(t, fullData, string(totalReadRunes), "expected total data")
}

func TestUnitRead_Empty_WithMin(t *testing.T) {
	ctx := context.Background()

	opts := &iorunez.Options{Min: 1}
	data := ""
	r := strings.NewReader(data)
	_, err := iorunez.Read(ctx, r, opts)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
}

func TestUnitRead_MultibyteCharacters(t *testing.T) {
	ctx := context.Background()
	data := "你好, world" // "Hello, world" in Chinese and English
	r := strings.NewReader(data)
	opts := &iorunez.Options{Max: 10}
	n, err := iorunez.Read(ctx, r, opts)
	require.NoError(t, err)
	assert.Equal(t, 9, n)
	assert.Equal(t, "你好, world", string(opts.Out))
}
