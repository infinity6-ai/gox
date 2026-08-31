package cryptzjwt_test

import (
	"testing"
	"time"

	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsakeyring"
	"github.com/stretchr/testify/assert"
)

func TestUnitSealUnsealOK(t *testing.T) {
	ts := time.UnixMilli(1770900600000).UTC()
	exp := time.Now().Add(10 * time.Minute).UTC()

	payload, err := cryptzjwt.NewToken().
		SetJti("myid").
		SetPayloadType("mytype").
		CreatedAt(ts).
		ExpiresAt(exp).
		SetIssuer("myiss").
		SetAudience("myaud").
		SetSubject("mysub").
		SetBothExtentions("strext", []byte("binext")).
		SetIdentityProvider("").
		BuildPayload()
	assert.NoError(t, err)

	otherKr := cryptzrsakeyring.NewKeyRing()
	otherKr.Generate("myaud", 1024)

	issKr := cryptzrsakeyring.NewKeyRing()
	issKr.Generate("myiss", 1024)

	header, tokenString := cryptzjwt.Seal(issKr, payload)
	assert.NotEmpty(t, tokenString)
	assert.Equal(t, &cryptzjwt.JWTHeader{
		Alg: "RS256",
		Typ: "JWT",
		Key: issKr.Current("myiss").KeyId(),
	}, header)

	token, err := cryptzjwt.UnsealKeyRing(otherKr, tokenString)
	assert.Error(t, err)
	assert.Empty(t, token)

	token, err = cryptzjwt.UnsealKeyRing(issKr, tokenString)
	assert.NoError(t, err)

	assert.Equal(t, &cryptzjwt.JWTToken{
		Header:  header,
		Payload: payload,
	}, token)

	err = cryptzjwt.NewVerifier().
		SetAudiences("myaud").
		SetIssuers(issKr.GetOwnerNames()...).
		SetBxts([]byte("binext")).SetExts("strext").
		Verify(token)

	assert.NoError(t, err)

}
