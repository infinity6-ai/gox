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

func assertCleanedOpts(t *testing.T, opts *iobytez.Options, expectedLen, expectedCap int) {
	t.Helper()

	opts.Clean()
	assert.Equal(t, expectedLen, len(opts.Out))
	assert.Equal(t, expectedCap, cap(opts.Out))
}

func TestUnitRead_Exact(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	r := bytes.NewReader([]byte("1234567890"))
	opts := &iobytez.Options{Min: 5, Max: 5}
	assertRead(t, ctx, r, opts, "12345", nil)

	assertCleanedOpts(t, opts, 0, 100)

	assertRead(t, ctx, r, opts, "67890", nil)

	assertCleanedOpts(t, opts, 0, 100)
	assertRead(t, ctx, r, opts, "", io.EOF)
}

func TestUnitRead_MinMax(t *testing.T) {
	t.Parallel()

	r := bytes.NewReader([]byte("1234567890"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 8,
	}
	assertRead(t, t.Context(), r, opts, "12345678", nil)
}

func TestUnitRead_UnexpectedEOF(t *testing.T) {
	t.Parallel()

	r := bytes.NewReader([]byte("123"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 10,
	}
	assertRead(t, t.Context(), r, opts, "123", io.ErrUnexpectedEOF)
}

func TestUnitRead_EOFSuccess(t *testing.T) {
	t.Parallel()

	r := bytes.NewReader([]byte("123456"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 10,
	}
	assertRead(t, t.Context(), r, opts, "123456", nil)
}

func TestUnitRead_EOFAtBeginning(t *testing.T) {
	t.Parallel()

	r := bytes.NewReader([]byte{})
	opts := &iobytez.Options{
		Min: 1,
		Max: 10,
	}
	assertRead(t, t.Context(), r, opts, "", io.EOF)
}

func TestUnitRead_NilOut(t *testing.T) {
	t.Parallel()

	r := bytes.NewReader([]byte("12345"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 5,
		Out: nil,
	}
	assertRead(t, t.Context(), r, opts, "12345", nil)
}
