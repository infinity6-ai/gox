package cryptzhash_test

import (
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzhash"
	"github.com/stretchr/testify/require"
)

func TestUnitSHA256(t *testing.T) {
	h, err := cryptzhash.SHA256Data("abc")
	require.NoError(t, err)
	require.Equal(t, "\xbax\x16\xbf\x8f\x01\xcf\xeaAA@\xde]\xae\"#\xb0\x03a\xa3\x96\x17z\x9c\xb4\x10\xffa\xf2\x00\x15\xad", string(h))
}

func TestUnitMD5(t *testing.T) {
	data := []byte("1-67-78")
	expectedHash := []byte{0xa1, 0x29, 0x20, 0xba, 0x37, 0x7e, 0xd8, 0xe7, 0x4d, 0xb9, 0xea, 0xda, 0x2c, 0x3f, 0x48, 0x61} // MD5 hash of "1-67-78"

	h, err := cryptzhash.MD5Data(data)
	require.NoError(t, err)
	require.Equal(t, expectedHash, h)

	h, err = cryptzhash.MD5Data("1-67-78")
	require.NoError(t, err)
	require.Equal(t, expectedHash, h)
}
