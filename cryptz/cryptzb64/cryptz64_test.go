package cryptzb64_test

import (
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
	"github.com/stretchr/testify/assert"
)

func TestUnitB64(t *testing.T) {
	data := []byte{0x73, 0xfc, 0x38, 0xfa}

	dec, err := cryptzb64.StdDecode("c/w4+g==")
	assert.NoError(t, err)
	assert.Equal(t, data, dec.Bytes())

	dec, err = cryptzb64.UrlDecode("c_w4-g")
	assert.NoError(t, err)
	assert.Equal(t, data, dec.Bytes())

	assert.Equal(t, "c/w4+g==", cryptzb64.StdEncode(data).String())
	assert.Equal(t, "c_w4-g", cryptzb64.UrlEncode(data).String())
}
