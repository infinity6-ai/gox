package cryptzjwtmsg_test

import (
	"crypto/rsa"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt/cryptzjwtmsg"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa"
	"github.com/stretchr/testify/assert"
)

func TestUnitSymEncryptDecrypt(t *testing.T) {
	s := cryptzrsa.NewService()
	fromKey := s.PrivKeyCreate(1024)
	destKey := s.PrivKeyCreate(1024)

	payloadSent, err := cryptzjwt.NewToken().
		GenerateJti().
		SetPayloadType("m").
		CreatedRightNow().
		ExpiresIn(10 * time.Minute).
		SetIssuer("src").
		SetAudience("dst").
		SetSubject("mysub").
		SetExtension("").
		SetIdentityProvider("").
		BuildPayload()

	assert.NoError(t, err)

	ciphertext := cryptzjwtmsg.Send(payloadSent, &cryptzjwtmsg.EncryptorKeysLoader{
		SrcKeyLoader: func(id string) *rsa.PrivateKey {
			assert.Equal(t, "src", id)
			return fromKey
		},
		DstKeyLoader: func(id string) *rsa.PublicKey {
			assert.Equal(t, "dst", id)
			return &destKey.PublicKey
		},
	})
	assert.NotNil(t, ciphertext)
	assert.Len(t, ciphertext, 376)

	payloadReceived := cryptzjwtmsg.Receive(ciphertext, &cryptzjwtmsg.DecryptorKeysLoader{
		SrcKeyLoader: func(id string) *rsa.PublicKey {
			assert.Equal(t, "src", id)
			return &fromKey.PublicKey
		},
		DstKeyLoader: func(id string) *rsa.PrivateKey {
			assert.Equal(t, "dst", id)
			return destKey
		},
	})

	assert.Equal(t, payloadSent, payloadReceived)
}
