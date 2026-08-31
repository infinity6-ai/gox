package cryptzb32_test

import (
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzb32"
	"github.com/stretchr/testify/require"
)

func TestUnitB32(t *testing.T) {
	b := []byte{0x73, 0xfc, 0x38, 0xfa}
	s := "efu3hug"
	require.Equal(t, s, cryptzb32.Encode(b).String())

	dec, err := cryptzb32.Decode(s)
	require.NoError(t, err)
	require.Equal(t, b, dec.Bytes())
}
