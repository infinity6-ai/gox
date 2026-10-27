package cryptz25519service_test

import (
	"context"
	"crypto/ed25519"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519service"
	"github.com/stretchr/testify/require"
)

func checkEncryptDecrypt(t *testing.T, sEncrypt, sDecrypt cryptz25519service.Service25519) {
	priv := sDecrypt.PrivKeyCreate()
	pub := priv.Public().(ed25519.PublicKey)

	crypted := sEncrypt.PubEncrypt(pub, []byte("aa"), []byte("mytext"))
	require.GreaterOrEqual(t, len(crypted), 32)
	require.Equal(t, "mytext", string(sDecrypt.PrivDecrypt(priv, []byte("aa"), crypted)))

	require.Panics(t, func() {
		sDecrypt.PrivDecrypt(priv, []byte("bb"), crypted)
	})
	require.NotEqual(t, crypted, sEncrypt.PubEncrypt(pub, []byte("aa"), []byte("mytext")))
}

func TestUnitEncryptDecrypt(t *testing.T) {
	services := getImplementations(context.Background())
	for _, s1 := range services { // Service to encrypt
		for _, s2 := range services { // Service to decrypt
			t.Run(buildName(s1, s2), func(t *testing.T) {
				checkEncryptDecrypt(t, s1, s2)
			})
		}
	}
}
