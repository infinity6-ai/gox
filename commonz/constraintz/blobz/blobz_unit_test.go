package blobz_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
)

func TestUnitWrapperStr(t *testing.T) {
	require.Equal(t, "hello", blobz.New("hello").String())
	require.Equal(t, []byte("hello"), blobz.New("hello").Bytes())

	require.True(t, blobz.New("hello").IsString())

	type My string
	require.Equal(t, "hello", blobz.New(My("hello")).String())
	require.Equal(t, []byte("hello"), blobz.New(My("hello")).Bytes())

	require.True(t, blobz.New(My("hello")).IsString())
}

func TestUnitWrapperBytes(t *testing.T) {
	require.Equal(t, "hello", blobz.New([]byte("hello")).String())
	require.Equal(t, []byte("hello"), blobz.New([]byte("hello")).Bytes())

	require.False(t, blobz.New([]byte("hello")).IsString())

	type My []byte
	require.Equal(t, "hello", blobz.New(My("hello")).String())
	require.Equal(t, []byte("hello"), blobz.New(My("hello")).Bytes())

	require.False(t, blobz.New(My("hello")).IsString())
}

func TestUnitToString(t *testing.T) {
	require.Equal(t, "hello", blobz.ToString("hello"))
	require.Equal(t, "hello", blobz.ToString([]byte("hello")))

	type MyString string
	require.Equal(t, "hello", blobz.ToString(MyString("hello")))

	type MyBytes []byte
	require.Equal(t, "hello", blobz.ToString(MyBytes("hello")))
}

func TestUnitToBytes(t *testing.T) {
	require.Equal(t, []byte("hello"), blobz.ToBytes("hello"))
	require.Equal(t, []byte("hello"), blobz.ToBytes([]byte("hello")))

	type MyString string
	require.Equal(t, []byte("hello"), blobz.ToBytes(MyString("hello")))

	type MyBytes []byte
	require.Equal(t, []byte("hello"), blobz.ToBytes(MyBytes("hello")))
}
