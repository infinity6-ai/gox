package cryptz25519impl

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/cryptz/cryptzaes"
	"github.com/infinity6-ai/gox/cryptz/cryptzhash"
)

type Service25519Impl struct{}

func New() *Service25519Impl {
	return &Service25519Impl{}
}

func (s *Service25519Impl) Name() string {
	return "cryptz25519impl"
}

func (s *Service25519Impl) PrivKeyCreate() ed25519.PrivateKey {
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	errorz.Check(err)
	return privKey
}

func (s *Service25519Impl) PrivKeyExport(privateKey ed25519.PrivateKey) string {
	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	errorz.Check(err)
	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyBytes,
	})
	return string(privKeyPEM)
}

func (s *Service25519Impl) PubKeyExport(publicKey ed25519.PublicKey) string {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	errorz.Check(err)
	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})
	return string(pubKeyPEM)
}

func (s *Service25519Impl) PubKeyId(pub ed25519.PublicKey) string {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pub)
	errorz.Check(err)
	hash, err := cryptzhash.SHA256Data(pubKeyBytes)
	errorz.Check(err)
	return fmt.Sprintf("%X", hash[len(hash)-8:])
}

func (s *Service25519Impl) PrivKeyImport(key string) ed25519.PrivateKey {
	block, _ := pem.Decode([]byte(key))
	if block == nil || block.Type != "PRIVATE KEY" {
		panic("failed to decode PEM block containing private key")
	}
	anykey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	errorz.Check(err)
	ret, ok := anykey.(ed25519.PrivateKey)
	if !ok {
		panic("not an Ed25519 private key")
	}
	return ret
}

func (s *Service25519Impl) PubKeyImport(key string) ed25519.PublicKey {
	block, _ := pem.Decode([]byte(key))
	if block == nil || block.Type != "PUBLIC KEY" {
		panic("failed to decode PEM block containing public key")
	}
	anypub, err := x509.ParsePKIXPublicKey(block.Bytes)
	errorz.Check(err)
	ret, ok := anypub.(ed25519.PublicKey)
	if !ok {
		panic("not an Ed25519 public key")
	}
	return ret
}

func (s *Service25519Impl) ExtractPubKey(privateKey ed25519.PrivateKey) ed25519.PublicKey {
	if len(privateKey) != ed25519.PrivateKeySize {
		panic("invalid ed25519 private key length")
	}
	return privateKey.Public().(ed25519.PublicKey)
}

func (s *Service25519Impl) Sign(privateKey ed25519.PrivateKey, message []byte) []byte {
	return ed25519.Sign(privateKey, message)
}

func (s *Service25519Impl) Verify(publicKey ed25519.PublicKey, message []byte, signature []byte) bool {
	return ed25519.Verify(publicKey, message, signature)
}

func (s *Service25519Impl) FindPubKey(pubs []ed25519.PublicKey, keyId string) ed25519.PublicKey {
	for _, pub := range pubs {
		if s.PubKeyId(pub) == keyId {
			return pub
		}
	}
	return nil
}

// X25519 derivation helpers

func (s *Service25519Impl) deriveX25519Priv(edPriv ed25519.PrivateKey) *ecdh.PrivateKey {
	seed := edPriv.Seed()
	h := sha512.Sum512(seed)
	clamped := make([]byte, 32)
	copy(clamped, h[:32])
	clamped[0] &= 248
	clamped[31] &= 127
	clamped[31] |= 64

	xPriv, err := ecdh.X25519().NewPrivateKey(clamped)
	errorz.Check(err)
	return xPriv
}

func (s *Service25519Impl) deriveX25519Pub(edPub ed25519.PublicKey) *ecdh.PublicKey {
	if len(edPub) != ed25519.PublicKeySize {
		panic("invalid Ed25519 public key size")
	}

	p, _ := new(big.Int).SetString("7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed", 16)

	yBytes := make([]byte, 32)
	copy(yBytes, edPub)
	yBytes[31] &= 0x7f

	rev := func(b []byte) []byte {
		r := make([]byte, len(b))
		for i := range b {
			r[len(b)-1-i] = b[i]
		}
		return r
	}
	y := new(big.Int).SetBytes(rev(yBytes))

	one := big.NewInt(1)
	num := new(big.Int).Add(one, y)
	num.Mod(num, p)

	den := new(big.Int).Sub(one, y)
	den.Mod(den, p)

	denInv := new(big.Int).ModInverse(den, p)
	u := new(big.Int).Mul(num, denInv)
	u.Mod(u, p)

	uBytesBig := u.Bytes()
	uBigPadded := make([]byte, 32)
	copy(uBigPadded[32-len(uBytesBig):], uBytesBig)
	uLE := rev(uBigPadded)

	xPub, err := ecdh.X25519().NewPublicKey(uLE)
	errorz.Check(err)
	return xPub
}

func (s *Service25519Impl) deriveAESKey(sharedSecret, ephPub, recipientPub, label []byte) []byte {
	hash := sha256.New()
	hash.Write(sharedSecret)
	hash.Write(ephPub)
	hash.Write(recipientPub)
	hash.Write(label)
	return hash.Sum(nil)
}

func (s *Service25519Impl) PubEncrypt(publicKey ed25519.PublicKey, label []byte, message []byte) []byte {
	recipientXPub := s.deriveX25519Pub(publicKey)
	ephPriv, err := ecdh.X25519().GenerateKey(rand.Reader)
	errorz.Check(err)
	ephPub := ephPriv.PublicKey()

	sharedSecret, err := ephPriv.ECDH(recipientXPub)
	errorz.Check(err)

	aesKey := s.deriveAESKey(sharedSecret, ephPub.Bytes(), recipientXPub.Bytes(), label)

	encData, err := cryptzaes.AESCrypt(cryptzaes.Key(aesKey), message)
	errorz.Check(err)

	// Payload is: EphemeralPubKey (32 bytes) || Ciphertext (including nonce & tag)
	payload := make([]byte, 0, 32+len(encData))
	payload = append(payload, ephPub.Bytes()...)
	payload = append(payload, encData...)

	return payload
}

func (s *Service25519Impl) PrivDecrypt(privateKey ed25519.PrivateKey, label []byte, payload []byte) []byte {
	if len(payload) < 32 {
		panic("payload too short, must contain at least 32 bytes for ephemeral public key")
	}

	ephPubBytes := payload[:32]
	ephPub, err := ecdh.X25519().NewPublicKey(ephPubBytes)
	errorz.Check(err)

	xPriv := s.deriveX25519Priv(privateKey)
	recipientXPub := xPriv.PublicKey()

	sharedSecret, err := xPriv.ECDH(ephPub)
	errorz.Check(err)

	aesKey := s.deriveAESKey(sharedSecret, ephPub.Bytes(), recipientXPub.Bytes(), label)

	plainblob, err := cryptzaes.AESDecrypt(cryptzaes.Key(aesKey), payload[32:])
	errorz.Check(err)

	return plainblob.Bytes()
}
