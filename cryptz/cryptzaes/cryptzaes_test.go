package cryptzaes_test

import (
	"slices"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzaes"
	"github.com/stretchr/testify/assert"
)

func TestUnitKey(t *testing.T) {
	assert.Len(t, cryptzaes.NewKey(32), 32)
	assert.Len(t, cryptzaes.NewKey(24), 24)
	assert.Len(t, cryptzaes.NewKey(16), 16)
	assert.Panics(t, func() { cryptzaes.NewKey(8) })
}

func tamper(idx int, data []byte) []byte {
	ret := append([]byte{}, data...)
	ret[idx] = ret[idx] + 1
	return ret
}

func TestUnitAESCryptAESDecrypt(t *testing.T) {
	key := cryptzaes.NewKey(16)

	plaintext := []byte("This is a test message.")
	ciphertext, err := cryptzaes.AESCrypt(key, plaintext)
	assert.NoError(t, err)

	original := slices.Clone(ciphertext)
	decrypted, err := cryptzaes.AESDecrypt(key, ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted.Bytes())
	assert.Equal(t, original, ciphertext)

	decrypted, err = cryptzaes.AESDecrypt(key, tamper(20, ciphertext))
	assert.Error(t, err)
	assert.Nil(t, decrypted)

	decrypted, err = cryptzaes.AESDecrypt(tamper(5, key), ciphertext)
	assert.Error(t, err)
	assert.Nil(t, decrypted)
}
