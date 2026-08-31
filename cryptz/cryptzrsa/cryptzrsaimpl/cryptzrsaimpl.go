package cryptzrsaimpl

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/validation"
	"github.com/infinity6-ai/gox/cryptz/cryptzhash"
)

type RSAServiceImpl struct{}

func New() *RSAServiceImpl {
	return &RSAServiceImpl{}
}

func (s *RSAServiceImpl) Name() string {
	return "cryptzrsaimpl"
}

func (s *RSAServiceImpl) PrivKeyCreate(bits int) *rsa.PrivateKey {
	validation.Greater(bits, 0, "bits")
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	errorz.Check(err)
	return privateKey
}

func (s *RSAServiceImpl) PrivKeyExport(privateKey *rsa.PrivateKey) string {
	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	errorz.Check(err)
	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyBytes,
	})
	return string(privKeyPEM)
}

func (s *RSAServiceImpl) PubKeyExport(publicKey *rsa.PublicKey) string {

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	errorz.Check(err)
	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})
	return string(pubKeyPEM)
}

func (s *RSAServiceImpl) PubKeyId(pub *rsa.PublicKey) string {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pub)
	errorz.Check(err)
	hash, err := cryptzhash.SHA256Data(pubKeyBytes)
	errorz.Check(err)
	return fmt.Sprintf("%X", hash[len(hash)-8:])
}

func (s *RSAServiceImpl) PrivKeyImport(key string) *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(key))
	if block == nil || block.Type != "PRIVATE KEY" {
		panic("failed to decode PEM block containing private key")
	}
	anykey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	errorz.Check(err)
	ret, ok := anykey.(*rsa.PrivateKey)
	if !ok {
		panic("not an RSA private key")
	}
	return ret
}

func (s *RSAServiceImpl) PubKeyImport(key string) *rsa.PublicKey {
	block, _ := pem.Decode([]byte(key))
	if block == nil || block.Type != "PUBLIC KEY" {
		panic("failed to decode PEM block containing public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	errorz.Check(err)
	return pub.(*rsa.PublicKey)
}

func (s *RSAServiceImpl) Sign(privateKey *rsa.PrivateKey, message []byte) []byte {
	// hash := sha256.New()
	// hash.Write(message)
	// hashed := hash.Sum(nil)
	hashed, err := cryptzhash.SHA256Data(message)
	errorz.Check(err)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	errorz.Check(err)
	return signature
}

func (s *RSAServiceImpl) Verify(publicKey *rsa.PublicKey, message []byte, signature []byte) bool {
	// hash := sha256.New()
	// hash.Write(message)
	// hashed := hash.Sum(nil)
	hashed, err := cryptzhash.SHA256Data(message)
	errorz.Check(err)
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed, signature)
	return err == nil
}

func (s *RSAServiceImpl) FindPubKey(pubs []*rsa.PublicKey, keyId string) *rsa.PublicKey {
	for _, pub := range pubs {
		if s.PubKeyId(pub) == keyId {
			return pub
		}
	}
	return nil
}

func (s *RSAServiceImpl) PubEncryptOEAP(publicKey *rsa.PublicKey, label []byte, message []byte) []byte {
	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, publicKey, message, label)
	errorz.Check(err)
	return ciphertext
}

func (s *RSAServiceImpl) PrivDecryptOEAP(privateKey *rsa.PrivateKey, label []byte, ciphertext []byte) []byte {
	hash := sha256.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, ciphertext, label)
	errorz.Check(err)
	return plaintext
}
