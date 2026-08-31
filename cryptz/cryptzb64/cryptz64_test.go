package cryptzb64_test

import (
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
	"github.com/stretchr/testify/require"
)

func TestUnitB64(t *testing.T) {
	data := []byte{0x73, 0xfc, 0x38, 0xfa}

	dec, err := cryptzb64.StdDecode("c/w4+g==")
	require.NoError(t, err)
	require.Equal(t, data, dec.Bytes())

	dec, err = cryptzb64.UrlDecode("c_w4-g")
	require.NoError(t, err)
	require.Equal(t, data, dec.Bytes())

	require.Equal(t, "c/w4+g==", cryptzb64.StdEncode(data).String())
	require.Equal(t, "c_w4-g", cryptzb64.UrlEncode(data).String())
}
