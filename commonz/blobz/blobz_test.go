package blobz_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.code.infinity6.ai/platform/util/blobz"
)

func TestUnitWrapperStr(t *testing.T) {
	assert.Equal(t, "hello", blobz.New("hello").String())
	assert.Equal(t, []byte("hello"), blobz.New("hello").Bytes())

	assert.True(t, blobz.New("hello").IsString())
	assert.False(t, blobz.New("hello").IsBytes())

	type My string
	assert.Equal(t, "hello", blobz.New(My("hello")).String())
	assert.Equal(t, []byte("hello"), blobz.New(My("hello")).Bytes())

	assert.True(t, blobz.New(My("hello")).IsString())
	assert.False(t, blobz.New(My("hello")).IsBytes())
}

func TestUnitWrapperBytes(t *testing.T) {
	assert.Equal(t, "hello", blobz.New([]byte("hello")).String())
	assert.Equal(t, []byte("hello"), blobz.New([]byte("hello")).Bytes())

	assert.False(t, blobz.New([]byte("hello")).IsString())
	assert.True(t, blobz.New([]byte("hello")).IsBytes())

	type My []byte
	assert.Equal(t, "hello", blobz.New(My("hello")).String())
	assert.Equal(t, []byte("hello"), blobz.New(My("hello")).Bytes())

	assert.False(t, blobz.New(My("hello")).IsString())
	assert.True(t, blobz.New(My("hello")).IsBytes())
}

func TestUnitToString(t *testing.T) {
	assert.Equal(t, "hello", blobz.ToString("hello"))
	assert.Equal(t, "hello", blobz.ToString([]byte("hello")))

	type MyString string
	assert.Equal(t, "hello", blobz.ToString(MyString("hello")))

	type MyBytes []byte
	assert.Equal(t, "hello", blobz.ToString(MyBytes("hello")))
}

func TestUnitToBytes(t *testing.T) {
	assert.Equal(t, []byte("hello"), blobz.ToBytes("hello"))
	assert.Equal(t, []byte("hello"), blobz.ToBytes([]byte("hello")))

	type MyString string
	assert.Equal(t, []byte("hello"), blobz.ToBytes(MyString("hello")))

	type MyBytes []byte
	assert.Equal(t, []byte("hello"), blobz.ToBytes(MyBytes("hello")))
}
