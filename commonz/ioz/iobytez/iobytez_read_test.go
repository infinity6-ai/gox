package iobytez_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/ioz/iobytez"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertRead(t *testing.T, ctx context.Context, r io.Reader, opts *iobytez.Options, expectedBuf string, expectedErr error) {
	err := iobytez.Read(ctx, r, opts)
	require.ErrorIs(t, err, expectedErr)
	assert.Equal(t, expectedBuf, string(opts.Out))
}

func TestUnitRead_Exact(t *testing.T) {
	ctx := t.Context()
	r := bytes.NewReader([]byte("1234567890"))
	opts := &iobytez.Options{Min: 5, Max: 5}
	assertRead(t, ctx, r, opts, "12345", nil)
}

func TestUnitRead_MinMax(t *testing.T) {
	r := bytes.NewReader([]byte("1234567890"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 8,
	}
	err := iobytez.Read(context.Background(), r, opts)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(opts.Out), 5)
	assert.LessOrEqual(t, len(opts.Out), 8)
}

func TestUnitRead_TimeoutSuccess(t *testing.T) {
	r := bytes.NewReader([]byte("1234567890"))
	opts := &iobytez.Options{
		Min:     5,
		Max:     5,
		Timeout: 100 * time.Millisecond,
	}
	err := iobytez.Read(context.Background(), r, opts)
	require.NoError(t, err)
	assert.Equal(t, 5, len(opts.Out))
	assert.Equal(t, []byte("12345"), opts.Out)
}

type blockingReader struct{}

func (r *blockingReader) Read(p []byte) (n int, err error) {
	time.Sleep(200 * time.Millisecond)
	return 0, nil
}

func TestUnitRead_TimeoutExceeded(t *testing.T) {
	r := &blockingReader{}
	opts := &iobytez.Options{
		Min:     1,
		Max:     10,
		Timeout: 100 * time.Millisecond,
	}
	err := iobytez.Read(context.Background(), r, opts)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestUnitRead_UnexpectedEOF(t *testing.T) {
	r := bytes.NewReader([]byte("123"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 10,
	}
	err := iobytez.Read(context.Background(), r, opts)
	require.Error(t, err)
	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
	assert.Equal(t, 3, len(opts.Out))
	assert.Equal(t, []byte("123"), opts.Out)
}

func TestUnitRead_EOFSuccess(t *testing.T) {
	r := bytes.NewReader([]byte("123456"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 10,
	}
	err := iobytez.Read(context.Background(), r, opts)
	require.NoError(t, err)
	assert.Equal(t, 6, len(opts.Out))
	assert.Equal(t, []byte("123456"), opts.Out)
}

func TestUnitRead_EOFAtBeginning(t *testing.T) {
	r := bytes.NewReader([]byte{})
	opts := &iobytez.Options{
		Min: 1,
		Max: 10,
	}
	err := iobytez.Read(context.Background(), r, opts)
	require.Error(t, err)
	assert.ErrorIs(t, err, io.EOF)
	assert.Equal(t, 0, len(opts.Out))
}

func TestUnitRead_NilOut(t *testing.T) {
	r := bytes.NewReader([]byte("12345"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 5,
		Out: nil,
	}
	err := iobytez.Read(context.Background(), r, opts)
	require.NoError(t, err)
	assert.Equal(t, 5, len(opts.Out))
	assert.Equal(t, []byte("12345"), opts.Out)
}
