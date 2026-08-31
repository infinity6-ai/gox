package cryptzrsaservice_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsaservice"
	"github.com/stretchr/testify/assert"
)

func checkEncryptDecryptOEAP(t *testing.T, sEncrypt, sDecrypt cryptzrsaservice.RSAService) {
	priv := sDecrypt.PrivKeyCreate(2048) // Create key with the decryption service
	pub := &priv.PublicKey

	crypted := sEncrypt.PubEncryptOEAP(pub, []byte("aa"), []byte("mytext"))
	assert.Equal(t, 256, len(crypted))
	assert.Equal(t, "mytext", string(sDecrypt.PrivDecryptOEAP(priv, []byte("aa"), crypted)))
	assert.Panics(t, func() {
		sDecrypt.PrivDecryptOEAP(priv, []byte("bb"), crypted)
	})
	assert.NotEqual(t, crypted, sEncrypt.PubEncryptOEAP(pub, []byte("aa"), []byte("mytext")))
}

func TestUnitEncryptDecryptOEAP(t *testing.T) {
	services := getImplementations(context.Background())
	for _, s1 := range services { // Service to encrypt
		for _, s2 := range services { // Service to decrypt
			t.Run(buildName(s1, s2), func(t *testing.T) {
				checkEncryptDecryptOEAP(t, s1, s2)
			})
		}
	}
}
