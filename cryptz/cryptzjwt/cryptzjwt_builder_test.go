package cryptzjwt_test

import (
	"testing"
	"time"

	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
	"github.com/stretchr/testify/assert"
)

func TestUnitStrictBuilder(t *testing.T) {
	creationTime := time.UnixMilli(1770900000000)
	expiresAt := time.Now().Add(10 * time.Hour)

	builder := cryptzjwt.NewToken().
		SetJti("myid").
		SetPayloadType("type").
		CreatedAt(creationTime).
		ExpiresAt(expiresAt).
		SetIssuer("issuer").
		SetAudience("audience").
		SetSubject("subject").
		SetBothExtentions("extension", []byte("mybxt")).
		SetIdentityProvider("idp")

	payload, err := builder.BuildPayload()
	assert.NoError(t, err)

	assert.Equal(t, &cryptzjwt.JWTPayload{
		Jti: "myid",
		Iss: "issuer",
		Sub: "subject",
		Iat: 1770900000000,
		Exp: expiresAt.UnixMilli(),
		Aud: "audience",
		Typ: "type",
		Ext: "extension",
		Bxt: []byte("mybxt"),
		Idp: "idp",
	}, payload)

	token, err := builder.SetAlg("myag").
		SetHeaderType("ht").
		SetKey("mykey").
		Build()
	assert.NoError(t, err)

	assert.Equal(t, &cryptzjwt.JWTHeader{
		Alg: "myag",
		Typ: "ht",
		Key: "mykey",
	}, token.Header)
	assert.Same(t, payload, token.Payload)
}

func TestUnitBuilderVariations(t *testing.T) {
	expirationTime := time.Now().Add(10 * time.Hour)

	builder := cryptzjwt.NewToken().
		GenerateJti().
		SetPayloadType("type").
		CreatedRightNow().
		ExpiresAt(expirationTime).
		SetIssuer("issuer").
		SetAudience("audience").
		SetSubject("subject").
		SetExtension("extension").
		SetIdentityProvider("idp")

	payload, err := builder.BuildPayload()
	assert.NoError(t, err)

	assert.NotEmpty(t, payload.Jti)
	assert.NotZero(t, payload.Iat)
	assert.LessOrEqual(t, payload.Iat, time.Now().UTC().UnixMilli())

	assert.Equal(t, &cryptzjwt.JWTPayload{
		Jti: payload.Jti,
		Iss: "issuer",
		Sub: "subject",
		Iat: payload.Iat,
		Exp: expirationTime.UTC().UnixMilli(),
		Aud: "audience",
		Typ: "type",
		Ext: "extension",
		Idp: "idp",
	}, payload)

	token, err := builder.SetAlg("myag").
		SetHeaderType("ht").
		SetKey("mykey").
		Build()
	assert.NoError(t, err)

	assert.Equal(t, &cryptzjwt.JWTHeader{
		Alg: "myag",
		Typ: "ht",
		Key: "mykey",
	}, token.Header)
	assert.Same(t, payload, token.Payload)
}
