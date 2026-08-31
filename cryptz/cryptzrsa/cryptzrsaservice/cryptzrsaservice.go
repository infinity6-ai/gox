package cryptzrsaservice

import (
	"crypto/rsa"
)

type RSAService interface {
	Name() string
	PrivKeyCreate(bits int) *rsa.PrivateKey
	PrivKeyExport(privateKey *rsa.PrivateKey) string
	PubKeyExport(publicKey *rsa.PublicKey) string
	PubKeyId(pub *rsa.PublicKey) string
	PrivKeyImport(key string) *rsa.PrivateKey
	PubKeyImport(key string) *rsa.PublicKey
	Sign(privateKey *rsa.PrivateKey, message []byte) []byte
	Verify(publicKey *rsa.PublicKey, message []byte, signature []byte) bool
	FindPubKey(pubs []*rsa.PublicKey, keyId string) *rsa.PublicKey
	PubEncryptOEAP(publicKey *rsa.PublicKey, label []byte, message []byte) []byte
	PrivDecryptOEAP(privateKey *rsa.PrivateKey, label []byte, ciphertext []byte) []byte
}
