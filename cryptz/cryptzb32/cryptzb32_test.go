package cryptzb32_test

import (
	"testing"

	"go.code.infinity6.ai/platform/cryptz/cryptzb32"

	"github.com/stretchr/testify/assert"
)

func TestUnitB32(t *testing.T) {
	b := []byte{0x73, 0xfc, 0x38, 0xfa}
	s := "efu3hug"
	assert.Equal(t, s, cryptzb32.Encode(b).String())
	assert.Equal(t, b, cryptzb32.Decode(s).Bytes())
}
