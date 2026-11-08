package cryptzjwt

import (
	"crypto/ed25519"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/validation"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/cryptz/cryptz25519"
	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519keyring"
	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
)

// Seal25519 seals a JWTPayload using an Ed25519 keyring.
func Seal25519(keyring *cryptz25519keyring.KeyRing, payload *JWTPayload) (*JWTHeader, string) {
	checker.NotNil(keyring, "keyring")
	checker.NotNil(payload, "payload")
	issuerKey := keyring.Current(payload.Iss)
	checker.NotNil(issuerKey, "issuerKey")
	checker.NotNil(issuerKey.Private(), "privateKey")

	tk := &JWTToken{
		Header: &JWTHeader{
			Alg: "EdDSA",
			Typ: "JWT",
			Key: issuerKey.KeyId(),
			Kid: issuerKey.KeyId(),
		},
		Payload: payload,
	}

	err := basicVerify(tk)
	errorz.Check(err)

	s := cryptz25519.NewService()

	unsignedToken := formatHeaderPayload(tk)
	sign := s.Sign(issuerKey.Private(), []byte(unsignedToken))
	signEncoded := cryptzb64.UrlEncode(sign).String()

	bundle := fmt.Sprintf("%s.%s", unsignedToken, signEncoded)
	return tk.Header, bundle
}

// Unseal25519Callback unseals an EdDSA JWT token using a public key resolver callback.
func Unseal25519Callback(token string, fn func(id string, iss string) (ed25519.PublicKey, error)) (*JWTToken, error) {
	s := cryptz25519.NewService()
	parsed, err := parseHeaderPayload(token)
	if err != nil {
		return nil, err
	}
	if parsed.Token.Header.Alg != "EdDSA" {
		return nil, fmt.Errorf("unsupported algorithm: %s, expected EdDSA", parsed.Token.Header.Alg)
	}
	key, err := fn(parsed.Token.Header.KeyId(), parsed.Token.Payload.Iss)
	if err != nil {
		return nil, err
	}
	if len(key) == 0 {
		return nil, fmt.Errorf("key not found")
	}
	ok := s.Verify(key, []byte(parsed.HeaderPayloadEncoded), parsed.Sign)
	if !ok {
		return nil, fmt.Errorf("wrong sign")
	}
	err = basicVerify(parsed.Token)
	return parsed.Token, err
}

// Unseal25519KeyRing unseals an EdDSA JWT token using an Ed25519 keyring.
func Unseal25519KeyRing(kr *cryptz25519keyring.KeyRing, token string) (*JWTToken, error) {
	return Unseal25519Callback(token, func(id, iss string) (ed25519.PublicKey, error) {
		key := kr.KeyIdOrNil(id)
		if key == nil {
			return nil, fmt.Errorf("key not found")
		}
		validation.Equal(iss, key.Owner(), "key id and owner iss mismatch")
		return key.Public(), nil
	})
}
