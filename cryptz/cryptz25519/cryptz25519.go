package cryptz25519

import (
	"context"
	"crypto/ed25519"
	"crypto/hkdf"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"fmt"

	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519impl"
	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519openssl"
	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519service"
)

// NewService returns a new Service25519 that uses the native Go crypto library.
func NewService() cryptz25519service.Service25519 {
	return cryptz25519impl.New()
}

// NewServiceOpenssl returns a new Service25519 that uses openssl commands.
func NewServiceOpenssl(ctx context.Context) cryptz25519service.Service25519 {
	return cryptz25519openssl.New(ctx)
}

// ExtractPubKey extracts the public key from the given ed25519.PrivateKey.
func ExtractPubKey(privateKey ed25519.PrivateKey) ed25519.PublicKey {
	return NewService().ExtractPubKey(privateKey)
}

// DeriveFromRSA deterministically derives an Ed25519 key pair from an existing RSA private key.
func DeriveFromRSA(rsaKey *rsa.PrivateKey) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	if rsaKey == nil {
		return nil, nil, fmt.Errorf("rsaKey cannot be nil")
	}

	rsaBytes := x509.MarshalPKCS1PrivateKey(rsaKey)

	seed, err := hkdf.Key(sha256.New, rsaBytes, []byte("cryptz:salt:v1"), "cryptz:rsa-to-ed25519:v1", ed25519.SeedSize)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to derive ed25519 seed: %w", err)
	}

	privKey := ed25519.NewKeyFromSeed(seed)
	pubKey := privKey.Public().(ed25519.PublicKey)

	return pubKey, privKey, nil
}

// DeriveEd25519FromRSA is an alias for DeriveFromRSA.
var DeriveEd25519FromRSA = DeriveFromRSA

