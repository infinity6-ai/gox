package iobytez_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/infinity6-ai/gox/commonz/ioz/iobytez"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertRead(t *testing.T, ctx context.Context, r io.Reader, opts *iobytez.Options, expectedBuf string, expectedErr error) {
	t.Helper()

	err := iobytez.Read(ctx, r, opts)
	require.ErrorIs(t, err, expectedErr)
	assert.Equal(t, expectedBuf, string(opts.Out))
}

func TestUnitRead_Exact(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	r := bytes.NewReader([]byte("1234567890"))
	opts := &iobytez.Options{Min: 5, Max: 5}
	assertRead(t, ctx, r, opts, "12345", nil)

	opts.Clean()
	assert.Equal(t, "", string(opts.Out))
	assert.Equal(t, 100, cap(opts.Out))
	assertRead(t, ctx, r, opts, "67890", nil)

	opts.Clean()
	assertRead(t, ctx, r, opts, "", io.EOF)
}

func TestUnitRead_MinMax(t *testing.T) {
	t.Parallel()

	r := bytes.NewReader([]byte("1234567890"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 8,
	}
	err := iobytez.Read(t.Context(), r, opts)
	require.NoError(t, err)
	assert.Len(t, opts.Out, 8)
}

func TestUnitRead_UnexpectedEOF(t *testing.T) {
	t.Parallel()

	r := bytes.NewReader([]byte("123"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 10,
	}
	err := iobytez.Read(t.Context(), r, opts)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
	assert.Equal(t, "123", string(opts.Out))
}

func TestUnitRead_EOFSuccess(t *testing.T) {
	t.Parallel()

	r := bytes.NewReader([]byte("123456"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 10,
	}
	err := iobytez.Read(t.Context(), r, opts)
	require.NoError(t, err)
	assert.Equal(t, "123456", string(opts.Out))
}

func TestUnitRead_EOFAtBeginning(t *testing.T) {
	t.Parallel()

	r := bytes.NewReader([]byte{})
	opts := &iobytez.Options{
		Min: 1,
		Max: 10,
	}
	err := iobytez.Read(t.Context(), r, opts)
	require.ErrorIs(t, err, io.EOF)
	assert.Empty(t, opts.Out)
}

func TestUnitRead_NilOut(t *testing.T) {
	t.Parallel()

	r := bytes.NewReader([]byte("12345"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 5,
		Out: nil,
	}
	err := iobytez.Read(t.Context(), r, opts)
	require.NoError(t, err)
	assert.Equal(t, "12345", string(opts.Out))
}
