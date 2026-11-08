package cryptzjwt_test

import (
	"crypto/ed25519"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519keyring"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
	"github.com/stretchr/testify/require"
)

func TestUnitSeal25519UnsealOK(t *testing.T) {
	ts := time.UnixMilli(1770900600000).UTC()
	exp := time.Now().Add(10 * time.Minute).UTC()

	payload, err := cryptzjwt.NewToken().
		SetJti("ed25519id").
		SetPayloadType("ed25519type").
		CreatedAt(ts).
		ExpiresAt(exp).
		SetIssuer("ediss").
		SetAudience("edaud").
		SetSubject("edsub").
		SetBothExtentions("strext", []byte("binext")).
		SetIdentityProvider("").
		BuildPayload()
	require.NoError(t, err)

	issKr := cryptz25519keyring.NewKeyRing()
	issKr.Generate("ediss")

	otherKr := cryptz25519keyring.NewKeyRing()
	otherKr.Generate("edaud")

	header, tokenString := cryptzjwt.Seal25519(issKr, payload)
	require.NotEmpty(t, tokenString)
	require.Equal(t, &cryptzjwt.JWTHeader{
		Alg: "EdDSA",
		Typ: "JWT",
		Kid: issKr.Current("ediss").KeyId(),
	}, header)

	// Unsealing with mismatched KeyRing fails
	token, err := cryptzjwt.Unseal25519KeyRing(otherKr, tokenString)
	require.Error(t, err)
	require.Nil(t, token)

	// Unsealing with matching KeyRing succeeds
	token, err = cryptzjwt.Unseal25519KeyRing(issKr, tokenString)
	require.NoError(t, err)
	require.Equal(t, &cryptzjwt.JWTToken{
		Header:  header,
		Payload: payload,
	}, token)

	// Verifier works on unsealed Ed25519 token
	err = cryptzjwt.NewVerifier().
		SetAudiences("edaud").
		SetIssuers(issKr.GetOwnerNames()...).
		SetBxts([]byte("binext")).
		SetExts("strext").
		SetHeaderAlgs("EdDSA").
		Verify(token)
	require.NoError(t, err)

	// Unseal25519Callback works
	cbToken, err := cryptzjwt.Unseal25519Callback(tokenString, func(id, iss string) (ed25519.PublicKey, error) {
		require.Equal(t, issKr.Current("ediss").KeyId(), id)
		require.Equal(t, "ediss", iss)
		return issKr.Current("ediss").Public(), nil
	})
	require.NoError(t, err)
	require.Equal(t, token, cbToken)
}
