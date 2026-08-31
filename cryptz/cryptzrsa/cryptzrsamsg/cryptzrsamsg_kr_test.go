package cryptzrsamsg_test

import (
	"testing"
	"time"

	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsakeyring"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsamsg"
	"github.com/stretchr/testify/assert"
)

func TestUnitMessageKR(t *testing.T) {
	kr := cryptzrsakeyring.NewKeyRing()
	kr.Generate("src", 1024)
	kr.Generate("dst", 1024)

	mkr := cryptzrsamsg.FromKeyRing(kr)

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

	ciphertext := cryptzrsamsg.Send(payloadSent, mkr.EncryptorKeysLoader())
	assert.NotNil(t, ciphertext)
	assert.Len(t, ciphertext, 376)

	payloadReceived := cryptzrsamsg.Receive(ciphertext, mkr.DecryptorKeysLoader())

	assert.Equal(t, payloadSent, payloadReceived)
}
