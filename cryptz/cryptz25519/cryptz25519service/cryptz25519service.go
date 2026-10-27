package cryptz25519service

import (
	"crypto/ed25519"
)

type Service25519 interface {
	Name() string
	PrivKeyCreate() ed25519.PrivateKey
	PrivKeyExport(privateKey ed25519.PrivateKey) string
	PubKeyExport(publicKey ed25519.PublicKey) string
	PubKeyId(pub ed25519.PublicKey) string
	PrivKeyImport(key string) ed25519.PrivateKey
	PubKeyImport(key string) ed25519.PublicKey
	ExtractPubKey(privateKey ed25519.PrivateKey) ed25519.PublicKey
	Sign(privateKey ed25519.PrivateKey, message []byte) []byte
	Verify(publicKey ed25519.PublicKey, message []byte, signature []byte) bool
	FindPubKey(pubs []ed25519.PublicKey, keyId string) ed25519.PublicKey
	PubEncrypt(publicKey ed25519.PublicKey, label []byte, message []byte) []byte
	PrivDecrypt(privateKey ed25519.PrivateKey, label []byte, ciphertext []byte) []byte
}
