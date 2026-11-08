package cryptzjwt

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsakeyring"
)

// SignRSA signs claims T using an RSA private key with RS256 and returns the RFC 7519 JWT string.
func SignRSA[T any](privKey *rsa.PrivateKey, keyId string, claims T) (string, error) {
	if privKey == nil {
		return "", fmt.Errorf("private key cannot be nil")
	}

	header := &JWTHeader{
		Alg: "RS256",
		Typ: "JWT",
		Kid: keyId,
	}

	headerJson := jsonz.MustFormat(header).Bytes()
	payloadJson := jsonz.MustFormat(claims).Bytes()

	headerEncoded := cryptzb64.UrlEncode(headerJson).String()
	payloadEncoded := cryptzb64.UrlEncode(payloadJson).String()
	toSign := fmt.Sprintf("%s.%s", headerEncoded, payloadEncoded)

	s := cryptzrsa.NewService()
	sig := s.Sign(privKey, []byte(toSign))
	sigEncoded := cryptzb64.UrlEncode(sig).String()

	return fmt.Sprintf("%s.%s", toSign, sigEncoded), nil
}

// VerifyRSA verifies an RS256 JWT using an RSA public key and decodes the claims into T.
func VerifyRSA[T any](pubKey *rsa.PublicKey, tokenStr string) (*JWTHeader, *T, error) {
	if pubKey == nil {
		return nil, nil, fmt.Errorf("public key cannot be nil")
	}

	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, nil, fmt.Errorf("invalid jwt format: expected 3 parts, got %d", len(parts))
	}

	headerJson, err := cryptzb64.UrlDecode(parts[0])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid header base64: %w", err)
	}
	payloadJson, err := cryptzb64.UrlDecode(parts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid payload base64: %w", err)
	}
	sig, err := cryptzb64.UrlDecode(parts[2])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid signature base64: %w", err)
	}

	s := cryptzrsa.NewService()
	toSign := fmt.Sprintf("%s.%s", parts[0], parts[1])
	if !s.Verify(pubKey, []byte(toSign), sig.Bytes()) {
		return nil, nil, fmt.Errorf("invalid signature")
	}

	var header JWTHeader
	if _, err := jsonz.Parse(headerJson.Bytes(), &header); err != nil {
		return nil, nil, fmt.Errorf("invalid header JSON: %w", err)
	}
	if header.Alg != "RS256" {
		return nil, nil, fmt.Errorf("unsupported algorithm: %s, expected RS256", header.Alg)
	}

	var claims T
	if _, err := jsonz.Parse(payloadJson.Bytes(), &claims); err != nil {
		return nil, nil, fmt.Errorf("invalid payload JSON: %w", err)
	}

	return &header, &claims, nil
}

// SealRSA signs claims T using the active key for the specified owner in the RSA keyring.
func SealRSA[T any](kr *cryptzrsakeyring.KeyRing, owner string, claims T) (*JWTHeader, string, error) {
	if kr == nil {
		return nil, "", fmt.Errorf("keyring cannot be nil")
	}
	key := kr.Current(owner)
	if key == nil {
		return nil, "", fmt.Errorf("key not found for owner %s", owner)
	}
	if key.Private() == nil {
		return nil, "", fmt.Errorf("private key not found for owner %s", owner)
	}

	header := &JWTHeader{
		Alg: "RS256",
		Typ: "JWT",
		Kid: key.KeyId(),
	}

	tokenStr, err := SignRSA(key.Private(), key.KeyId(), claims)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign token: %w", err)
	}
	return header, tokenStr, nil
}

// UnsealRSACallback verifies and decodes an RS256 token using a key-lookup callback.
func UnsealRSACallback[T any](tokenStr string, fn func(keyId, iss string) (*rsa.PublicKey, error)) (*JWTHeader, *T, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, nil, fmt.Errorf("invalid jwt format: expected 3 parts, got %d", len(parts))
	}

	headerJson, err := cryptzb64.UrlDecode(parts[0])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid header base64: %w", err)
	}
	payloadJson, err := cryptzb64.UrlDecode(parts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid payload base64: %w", err)
	}
	sig, err := cryptzb64.UrlDecode(parts[2])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid signature base64: %w", err)
	}

	var header JWTHeader
	if _, err := jsonz.Parse(headerJson.Bytes(), &header); err != nil {
		return nil, nil, fmt.Errorf("invalid header JSON: %w", err)
	}
	if header.Alg != "RS256" {
		return nil, nil, fmt.Errorf("unsupported algorithm: %s, expected RS256", header.Alg)
	}

	var probe struct {
		Iss string `json:"iss"`
	}
	_ = json.Unmarshal(payloadJson.Bytes(), &probe)

	pubKey, err := fn(header.KeyId(), probe.Iss)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve public key: %w", err)
	}
	if pubKey == nil {
		return nil, nil, fmt.Errorf("public key not found for keyId=%s iss=%s", header.KeyId(), probe.Iss)
	}

	s := cryptzrsa.NewService()
	toSign := fmt.Sprintf("%s.%s", parts[0], parts[1])
	if !s.Verify(pubKey, []byte(toSign), sig.Bytes()) {
		return nil, nil, fmt.Errorf("invalid signature")
	}

	var claims T
	if _, err := jsonz.Parse(payloadJson.Bytes(), &claims); err != nil {
		return nil, nil, fmt.Errorf("invalid payload JSON: %w", err)
	}

	return &header, &claims, nil
}

// UnsealRSA verifies and decodes an RS256 token using an RSA keyring.
func UnsealRSA[T any](kr *cryptzrsakeyring.KeyRing, tokenStr string) (*JWTHeader, *T, error) {
	if kr == nil {
		return nil, nil, fmt.Errorf("keyring cannot be nil")
	}
	return UnsealRSACallback[T](tokenStr, func(keyId, iss string) (*rsa.PublicKey, error) {
		key := kr.KeyIdOrNil(keyId)
		if key == nil {
			return nil, fmt.Errorf("key not found for id=%s", keyId)
		}
		if iss != "" && key.Owner() != iss {
			return nil, fmt.Errorf("key owner %s does not match issuer %s", key.Owner(), iss)
		}
		return key.Public(), nil
	})
}
