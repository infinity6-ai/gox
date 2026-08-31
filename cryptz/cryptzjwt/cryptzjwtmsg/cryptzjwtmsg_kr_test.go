package cryptzjwtmsg_test

import (
	"testing"
	"time"

	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt/cryptzjwtmsg"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsakeyring"
	"github.com/stretchr/testify/assert"
)

func TestUnitMessageKR(t *testing.T) {
	kr := cryptzrsakeyring.NewKeyRing()
	kr.Generate("src", 1024)
	kr.Generate("dst", 1024)

	mkr := cryptzjwtmsg.FromKeyRing(kr)

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

	ciphertext := cryptzjwtmsg.Send(payloadSent, mkr.EncryptorKeysLoader())
	assert.NotNil(t, ciphertext)
	assert.Len(t, ciphertext, 376)

	payloadReceived := cryptzjwtmsg.Receive(ciphertext, mkr.DecryptorKeysLoader())

	assert.Equal(t, payloadSent, payloadReceived)
}
