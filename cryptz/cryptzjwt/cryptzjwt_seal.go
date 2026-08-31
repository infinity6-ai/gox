package cryptzjwt

import (
	"crypto/rsa"
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/validation"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsakeyring"
	"go.code.infinity6.ai/platform/util/jsonz"
)

func Seal(keyring *cryptzrsakeyring.KeyRing, payload *JWTPayload) (*JWTHeader, string) {

	issuerKey := keyring.Current(payload.Iss)
	tk := &JWTToken{
		Header: &JWTHeader{
			Alg: "RS256",
			Typ: "JWT",
			Key: issuerKey.KeyId(),
		},
		Payload: payload,
	}

	err := basicVerify(tk)
	errorz.Check(err)

	s := cryptzrsa.NewService()

	unsignedToken := formatHeaderPayload(tk)
	sign := s.Sign(issuerKey.Private(), []byte(unsignedToken))
	signEncoded := cryptzb64.UrlEncode(sign).String()

	bundle := fmt.Sprintf("%s.%s", unsignedToken, signEncoded)
	return tk.Header, bundle
}

func basicVerify(tk *JWTToken) error {
	return NewVerifier().
		AnyAudience().
		AnyIssuers().
		Verify(tk)
}

func formatHeaderPayload(tk *JWTToken) string {
	headerJson := jsonz.FormatString(tk.Header)
	payloadJson := jsonz.FormatString(tk.Payload)
	headerEncoded := cryptzb64.UrlEncode(headerJson).String()
	payloadEncoded := cryptzb64.UrlEncode(payloadJson).String()
	unsignedToken := fmt.Sprintf("%s.%s", headerEncoded, payloadEncoded)
	return unsignedToken
}

type JWTParsedToken struct {
	Token                *JWTToken
	HeaderPayloadEncoded string
	Sign                 []byte
}

func parseHeaderPayload(token string) (*JWTParsedToken, error) {
	parts := strings.Split(token, ".")
	checker.Equal(3, len(parts), "parts")

	headerEncoded := parts[0]
	payloadEncoded := parts[1]
	signEncoded := parts[2]

	headerJson, err := cryptzb64.UrlDecode(headerEncoded)
	if err != nil {
		return nil, err
	}
	payloadJson, err := cryptzb64.UrlDecode(payloadEncoded)
	if err != nil {
		return nil, err
	}

	headerParsed, err := jsonz.ParseE(headerJson.Bytes(), &JWTHeader{})
	if err != nil {
		return nil, err
	}

	payloadParsed, err := jsonz.ParseE(payloadJson.Bytes(), &JWTPayload{})
	if err != nil {
		return nil, err
	}

	tk := &JWTToken{
		Header:  headerParsed,
		Payload: payloadParsed,
	}

	sign, err := cryptzb64.UrlDecode(signEncoded)
	if err != nil {
		return nil, err
	}

	ret := &JWTParsedToken{
		Token:                tk,
		HeaderPayloadEncoded: fmt.Sprintf("%s.%s", headerEncoded, payloadEncoded),
		Sign:                 sign.Bytes(),
	}

	return ret, nil
}

func UnsealCallback(token string, fn func(id string, iss string) (*rsa.PublicKey, error)) (*JWTToken, error) {
	s := cryptzrsa.NewService()
	parsed, err := parseHeaderPayload(token)
	if err != nil {
		return nil, err
	}
	key, err := fn(parsed.Token.Header.Key, parsed.Token.Payload.Iss)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, fmt.Errorf("key not found")
	}
	ok := s.Verify(key, []byte(parsed.HeaderPayloadEncoded), parsed.Sign)
	if !ok {
		return nil, fmt.Errorf("wrong sign")
	}
	err = basicVerify(parsed.Token)
	return parsed.Token, err
}

func UnsealKeyRing(kr *cryptzrsakeyring.KeyRing, token string) (*JWTToken, error) {
	return UnsealCallback(token, func(id, iss string) (*rsa.PublicKey, error) {
		key := kr.KeyIdOrNil(id)
		if key == nil {
			return nil, fmt.Errorf("key not found")
		}
		validation.Equal(iss, key.Owner(), "key id and owner iss mismatch")
		return key.Public(), nil
	})
}
