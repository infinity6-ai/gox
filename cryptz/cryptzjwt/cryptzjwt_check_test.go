package cryptzjwt_test

import (
	"crypto/rsa"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsakeyring"
	"github.com/stretchr/testify/assert"
)

func checkSign(t *testing.T, token string, pub *rsa.PublicKey) {
	_, err := cryptzjwt.UnsealCallback(token+"x", func(id, iss string) (*rsa.PublicKey, error) {
		return pub, nil
	})
	assert.ErrorContains(t, err, "wrong sign")

	_, err = cryptzjwt.UnsealCallback(token, func(id, iss string) (*rsa.PublicKey, error) {
		return &cryptzrsa.NewService().PrivKeyCreate(1024).PublicKey, nil
	})
	assert.ErrorContains(t, err, "wrong sign")

}

func checkIss(t *testing.T, token string, pub *rsa.PublicKey) {
	_, err := cryptzjwt.UnsealCallback(token, func(id, iss string) (*rsa.PublicKey, error) {
		return &cryptzrsa.NewService().PrivKeyCreate(1024).PublicKey, nil
	})
	assert.ErrorContains(t, err, "wrong sign")

	_, err = cryptzjwt.UnsealCallback(token, func(id, iss string) (*rsa.PublicKey, error) {
		return pub, nil
	})
	assert.NoError(t, err)

}

func checkAud(t *testing.T, token string, pub *rsa.PublicKey) {
	parsedToken, err := cryptzjwt.UnsealCallback(token, func(id, iss string) (*rsa.PublicKey, error) {
		return pub, nil
	})
	assert.NoError(t, err)
	err = cryptzjwt.NewVerifier().SetAudiences("other").AnyIssuers().Verify(parsedToken)
	assert.ErrorContains(t, err, "audience not allowed")

	parsedToken, err = cryptzjwt.UnsealCallback(token, func(id, iss string) (*rsa.PublicKey, error) {
		return pub, nil
	})
	assert.NoError(t, err)
	err = cryptzjwt.NewVerifier().SetAudiences("a").AnyIssuers().Verify(parsedToken)
	assert.NoError(t, err)

}

func checkCreateRequiredFields(t *testing.T) {
	p, err := cryptzjwt.NewToken().GenerateJti().SetPayloadType("t").
		CreatedAt(time.Now()).ExpiresAt(time.Now().Add(10 * time.Second)).
		SetIssuer("i").SetAudience("a").
		SetSubject("abc").SetExtension("").SetIdentityProvider("").
		BuildPayload()
	assert.NoError(t, err)
	cryptzjwt.NewVerifier().AnyAudience().AnyIssuers().MustVerify(&cryptzjwt.JWTToken{Payload: p})

	_, err = cryptzjwt.NewToken().GenerateJti().SetPayloadType("t").
		CreatedAt(time.UnixMilli(0).UTC()).ExpiresAt(time.Now().Add(10 * time.Second)).
		SetIssuer("i").SetAudience("a").
		SetSubject("abc").SetExtension("").SetIdentityProvider("").
		BuildPayload()
	assert.ErrorContains(t, err, "iat is required")

	_, err = cryptzjwt.NewToken().GenerateJti().SetPayloadType("t").
		CreatedAt(time.Now()).ExpiresAt(time.Now().Add(10 * time.Second)).
		SetIssuer("").SetAudience("a").
		SetSubject("abc").SetExtension("").SetIdentityProvider("").
		BuildPayload()
	assert.ErrorContains(t, err, "issuer is required")

	_, err = cryptzjwt.NewToken().GenerateJti().SetPayloadType("t").
		CreatedAt(time.Now()).ExpiresAt(time.Now().Add(10 * time.Second)).
		SetIssuer("i").SetAudience("").
		SetSubject("abc").SetExtension("").SetIdentityProvider("").
		BuildPayload()
	assert.ErrorContains(t, err, "audience is required")

	_, err = cryptzjwt.NewToken().GenerateJti().SetPayloadType("t").
		CreatedAt(time.Now()).ExpiresAt(time.Now().Add(10 * time.Second)).
		SetIssuer("i").SetAudience("a").
		SetSubject("").SetExtension("").SetIdentityProvider("").
		BuildPayload()
	assert.ErrorContains(t, err, "sub is required")

	_, err = cryptzjwt.NewToken().GenerateJti().SetPayloadType("t").
		CreatedAt(time.Now()).ExpiresAtEpoch(0).
		SetIssuer("i").SetAudience("a").
		SetSubject("abc").SetExtension("").SetIdentityProvider("").
		BuildPayload()
	assert.ErrorContains(t, err, "exp is required")

}

func checkCreateFutureToken(t *testing.T) {
	_, err := cryptzjwt.NewToken().GenerateJti().SetPayloadType("t").
		CreatedAt(time.Now().Add(10 * time.Hour)).ExpiresAt(time.Now().Add(11 * time.Hour)).
		SetIssuer("i").SetAudience("a").
		SetSubject("abc").SetExtension("").SetIdentityProvider("").
		BuildPayload()
	assert.ErrorContains(t, err, "iat is in the future")

}

func TestUnitJWTBasic(t *testing.T) {
	kr := cryptzrsakeyring.NewKeyRing()
	kr.Generate("i", 1024)

	checkCreateRequiredFields(t)
	checkCreateFutureToken(t)

	p, err := cryptzjwt.NewToken().GenerateJti().SetPayloadType("t").
		CreatedAt(time.Now()).ExpiresIn(10 * time.Second).
		SetIssuer("i").SetAudience("a").
		SetSubject("abc").SetExtension("").SetIdentityProvider("").
		BuildPayload()
	assert.NoError(t, err)
	_, token := cryptzjwt.Seal(kr, p)

	assert.Regexp(t, "^[A-Za-z0-9_\\-]+\\.[A-Za-z0-9_\\-]+\\.[A-Za-z0-9_\\-]+$", token)

	parsedToken, err := cryptzjwt.UnsealKeyRing(kr, token)
	assert.NoError(t, err)

	assert.NotEmpty(t, parsedToken.Payload.Jti)
	assert.Equal(t, "abc", parsedToken.Payload.Sub)
	assert.Equal(t, "RS256", parsedToken.Header.Alg)
	assert.Equal(t, "JWT", parsedToken.Header.Typ)
	assert.Equal(t, kr.Current("i").KeyId(), parsedToken.Header.Key)

	checkSign(t, token, kr.Current("i").Public())

	checkIss(t, token, kr.Current("i").Public())
	checkAud(t, token, kr.Current("i").Public())
}

func TestUnitJWTExp(t *testing.T) {
	kr := cryptzrsakeyring.NewKeyRing()
	kr.Generate("i", 1024)

	p, err := cryptzjwt.NewToken().GenerateJti().SetPayloadType("t").
		CreatedAt(time.Now()).ExpiresIn(250 * time.Millisecond).
		SetIssuer("i").SetAudience("a").
		SetSubject("abc").SetExtension("").SetIdentityProvider("").
		BuildPayload()
	assert.NoError(t, err)
	h, token := cryptzjwt.Seal(kr, p)

	parsedToken, err := cryptzjwt.UnsealKeyRing(kr, token)
	assert.NoError(t, err)
	assert.NotNil(t, parsedToken.Header)
	assert.NotNil(t, parsedToken.Payload)
	assert.Equal(t, h, parsedToken.Header)

	err = cryptzjwt.NewVerifier().SetAudiences("a").SetIssuers("i").Verify(parsedToken)
	assert.NoError(t, err)

	diff := parsedToken.Payload.Exp - time.Now().UTC().UnixMilli() // datez.NowMilliInt()
	assert.Greater(t, diff, int64(0))
	// util.Sleep(diff + 1)
	time.Sleep(time.Duration(diff+1) * time.Millisecond)

	err = cryptzjwt.NewVerifier().SetAudiences("a").SetIssuers("i").Verify(parsedToken)
	assert.ErrorContains(t, err, "expired")

	parsedToken, err = cryptzjwt.UnsealKeyRing(kr, token)
	assert.ErrorContains(t, err, "expired")

	err = cryptzjwt.NewVerifier().SetAudiences("a").SetIssuers("i").Verify(parsedToken)
	assert.ErrorContains(t, err, "expired")
}
