package cryptzaes_test

import (
	"slices"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzaes"
	"github.com/stretchr/testify/require"
)

func TestUnitKey(t *testing.T) {
	require.Len(t, cryptzaes.NewKey(32), 32)
	require.Len(t, cryptzaes.NewKey(24), 24)
	require.Len(t, cryptzaes.NewKey(16), 16)
	require.Panics(t, func() { cryptzaes.NewKey(8) })
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
	require.NoError(t, err)

	original := slices.Clone(ciphertext)
	decrypted, err := cryptzaes.AESDecrypt(key, ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted.Bytes())
	require.Equal(t, original, ciphertext)

	decrypted, err = cryptzaes.AESDecrypt(key, tamper(20, ciphertext))
	require.Error(t, err)
	require.Nil(t, decrypted)

	decrypted, err = cryptzaes.AESDecrypt(tamper(5, key), ciphertext)
	require.Error(t, err)
	require.Nil(t, decrypted)
}
