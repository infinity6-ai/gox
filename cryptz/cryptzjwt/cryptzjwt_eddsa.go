package cryptzjwt

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/cryptz/cryptz25519"
	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519keyring"
	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
)

// StandardClaims represents standard RFC 7519 JWT claims.
type StandardClaims struct {
	Jti string `json:"jti,omitempty"`
	Iss string `json:"iss,omitempty"`
	Sub string `json:"sub,omitempty"`
	Aud string `json:"aud,omitempty"`
	Iat int64  `json:"iat,omitempty"`
	Nbf int64  `json:"nbf,omitempty"`
	Exp int64  `json:"exp,omitempty"`
	Typ string `json:"typ,omitempty"`
}

// SignEdDSA signs claims T using an Ed25519 private key and returns the RFC 8037 JWT string.
func SignEdDSA[T any](privKey ed25519.PrivateKey, keyId string, claims T) (string, error) {
	if len(privKey) == 0 {
		return "", fmt.Errorf("private key cannot be empty")
	}

	header := &JWTHeader{
		Alg: "EdDSA",
		Typ: "JWT",
		Kid: keyId,
	}

	headerJson := jsonz.MustFormat(header).Bytes()
	payloadJson := jsonz.MustFormat(claims).Bytes()

	headerEncoded := cryptzb64.UrlEncode(headerJson).String()
	payloadEncoded := cryptzb64.UrlEncode(payloadJson).String()
	toSign := fmt.Sprintf("%s.%s", headerEncoded, payloadEncoded)

	s := cryptz25519.NewService()
	sig := s.Sign(privKey, []byte(toSign))
	sigEncoded := cryptzb64.UrlEncode(sig).String()

	return fmt.Sprintf("%s.%s", toSign, sigEncoded), nil
}

// VerifyEdDSA verifies an EdDSA JWT using an Ed25519 public key and decodes the claims into T.
func VerifyEdDSA[T any](pubKey ed25519.PublicKey, tokenStr string) (*JWTHeader, *T, error) {
	if len(pubKey) == 0 {
		return nil, nil, fmt.Errorf("public key cannot be empty")
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

	s := cryptz25519.NewService()
	toSign := fmt.Sprintf("%s.%s", parts[0], parts[1])
	if !s.Verify(pubKey, []byte(toSign), sig.Bytes()) {
		return nil, nil, fmt.Errorf("invalid signature")
	}

	var header JWTHeader
	if _, err := jsonz.Parse(headerJson.Bytes(), &header); err != nil {
		return nil, nil, fmt.Errorf("invalid header JSON: %w", err)
	}
	if header.Alg != "EdDSA" {
		return nil, nil, fmt.Errorf("unsupported algorithm: %s, expected EdDSA", header.Alg)
	}

	var claims T
	if _, err := jsonz.Parse(payloadJson.Bytes(), &claims); err != nil {
		return nil, nil, fmt.Errorf("invalid payload JSON: %w", err)
	}

	return &header, &claims, nil
}

// SealEdDSA signs claims T using the active key for the specified owner in the Ed25519 keyring.
func SealEdDSA[T any](kr *cryptz25519keyring.KeyRing, owner string, claims T) (*JWTHeader, string, error) {
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
		Alg: "EdDSA",
		Typ: "JWT",
		Kid: key.KeyId(),
	}

	tokenStr, err := SignEdDSA(key.Private(), key.KeyId(), claims)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign token: %w", err)
	}
	return header, tokenStr, nil
}

// UnsealEdDSACallback verifies and decodes an EdDSA token using a key-lookup callback.
func UnsealEdDSACallback[T any](tokenStr string, fn func(keyId, iss string) (ed25519.PublicKey, error)) (*JWTHeader, *T, error) {
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
	if header.Alg != "EdDSA" {
		return nil, nil, fmt.Errorf("unsupported algorithm: %s, expected EdDSA", header.Alg)
	}

	var probe struct {
		Iss string `json:"iss"`
	}
	_ = json.Unmarshal(payloadJson.Bytes(), &probe)

	pubKey, err := fn(header.KeyId(), probe.Iss)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve public key: %w", err)
	}
	if len(pubKey) == 0 {
		return nil, nil, fmt.Errorf("public key not found for keyId=%s iss=%s", header.KeyId(), probe.Iss)
	}

	s := cryptz25519.NewService()
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

// UnsealEdDSA verifies and decodes an EdDSA token using an Ed25519 keyring.
func UnsealEdDSA[T any](kr *cryptz25519keyring.KeyRing, tokenStr string) (*JWTHeader, *T, error) {
	if kr == nil {
		return nil, nil, fmt.Errorf("keyring cannot be nil")
	}
	return UnsealEdDSACallback[T](tokenStr, func(keyId, iss string) (ed25519.PublicKey, error) {
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
