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

func TestRead_Exact(t *testing.T) {
	r := bytes.NewReader([]byte("1234567890"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 5,
	}
	err := iobytez.Read(context.Background(), r, opts)
	require.NoError(t, err)
	assert.Equal(t, 5, opts.Len)
	assert.Equal(t, []byte("12345"), opts.Out)
}

func TestRead_MinMax(t *testing.T) {
	r := bytes.NewReader([]byte("1234567890"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 8,
	}
	err := iobytez.Read(context.Background(), r, opts)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, opts.Len, 5)
	assert.LessOrEqual(t, opts.Len, 8)
}

func TestRead_TimeoutSuccess(t *testing.T) {
	r := bytes.NewReader([]byte("1234567890"))
	opts := &iobytez.Options{
		Min:     5,
		Max:     5,
		Timeout: 100 * time.Millisecond,
	}
	err := iobytez.Read(context.Background(), r, opts)
	require.NoError(t, err)
	assert.Equal(t, 5, opts.Len)
	assert.Equal(t, []byte("12345"), opts.Out)
}

type blockingReader struct{}

func (r *blockingReader) Read(p []byte) (n int, err error) {
	time.Sleep(200 * time.Millisecond)
	return 0, nil
}

func TestRead_TimeoutExceeded(t *testing.T) {
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

func TestRead_UnexpectedEOF(t *testing.T) {
	r := bytes.NewReader([]byte("123"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 10,
	}
	err := iobytez.Read(context.Background(), r, opts)
	require.Error(t, err)
	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
	assert.Equal(t, 3, opts.Len)
	assert.Equal(t, []byte("123"), opts.Out)
}

func TestRead_EOFSuccess(t *testing.T) {
	r := bytes.NewReader([]byte("123456"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 10,
	}
	err := iobytez.Read(context.Background(), r, opts)
	require.NoError(t, err)
	assert.Equal(t, 6, opts.Len)
	assert.Equal(t, []byte("123456"), opts.Out)
}

func TestRead_EOFAtBeginning(t *testing.T) {
	r := bytes.NewReader([]byte{})
	opts := &iobytez.Options{
		Min: 1,
		Max: 10,
	}
	err := iobytez.Read(context.Background(), r, opts)
	require.Error(t, err)
	assert.ErrorIs(t, err, io.EOF)
	assert.Equal(t, 0, opts.Len)
}

func TestRead_NilOut(t *testing.T) {
	r := bytes.NewReader([]byte("12345"))
	opts := &iobytez.Options{
		Min: 5,
		Max: 5,
		Out: nil,
	}
	err := iobytez.Read(context.Background(), r, opts)
	require.NoError(t, err)
	assert.Equal(t, 5, opts.Len)
	assert.Equal(t, []byte("12345"), opts.Out)
}
