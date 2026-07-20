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

func TestUnitRead_Incremental(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	r := bytes.NewReader([]byte("1234567890"))
	// Start with non-empty buffer, it should be appended to.
	opts := &iobytez.Options{
		Out: []byte("initial"),
	}

	// Read 1: Min 2, Max 4. Should read "1234".
	opts.Min = 2
	opts.Max = 4
	assertRead(t, ctx, r, opts, "initial1234", nil)

	// Read 2: Min 2, Max 4. Should read "5678".
	opts.Min = 2
	opts.Max = 4
	assertRead(t, ctx, r, opts, "initial12345678", nil)

	// Read 3: Min 1, Max 4. Should read "90" and hit EOF. Read >= Min, so no error.
	opts.Min = 1
	opts.Max = 4
	assertRead(t, ctx, r, opts, "initial1234567890", nil)

	// Read 4: At EOF. Min 1, Max 4. Should get ErrUnexpectedEOF because we can't read Min.
	opts.Min = 1
	opts.Max = 4
	assertRead(t, ctx, r, opts, "initial1234567890", io.ErrUnexpectedEOF)

	// Read 5: At EOF. Min 0, Max 4. Should get EOF because we read 0 bytes.
	opts.Min = 0
	opts.Max = 4
	assertRead(t, ctx, r, opts, "initial1234567890", io.EOF)
}
