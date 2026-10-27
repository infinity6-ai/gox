package cryptz25519openssl

import (
	"context"
	"crypto/ed25519"
	"os"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/execz"
	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519impl"
)

type Service25519Openssl struct {
	impl *cryptz25519impl.Service25519Impl
	ctx  context.Context
}

func New(ctx context.Context) *Service25519Openssl {
	return &Service25519Openssl{
		impl: cryptz25519impl.New(),
		ctx:  ctx,
	}
}

func (s *Service25519Openssl) Name() string {
	return "cryptz25519openssl"
}

func (s *Service25519Openssl) PrivKeyCreate() ed25519.PrivateKey {
	privKeyFilePath := filez.CreateTempFile("private_key.pem", nil)
	defer os.Remove(privKeyFilePath)

	err := execz.New(s.ctx, "openssl", "genpkey", "-algorithm", "ED25519", "-out", privKeyFilePath).Run()
	errorz.Check(err)

	keyData := filez.MustReadFile(privKeyFilePath, 4096)
	return s.PrivKeyImport(keyData.String())
}

func (s *Service25519Openssl) PrivKeyExport(privateKey ed25519.PrivateKey) string {
	return s.impl.PrivKeyExport(privateKey)
}

func (s *Service25519Openssl) PubKeyExport(publicKey ed25519.PublicKey) string {
	return s.impl.PubKeyExport(publicKey)
}

func (s *Service25519Openssl) PubKeyId(pub ed25519.PublicKey) string {
	return s.impl.PubKeyId(pub)
}

func (s *Service25519Openssl) PrivKeyImport(key string) ed25519.PrivateKey {
	return s.impl.PrivKeyImport(key)
}

func (s *Service25519Openssl) PubKeyImport(key string) ed25519.PublicKey {
	return s.impl.PubKeyImport(key)
}

func (s *Service25519Openssl) ExtractPubKey(privateKey ed25519.PrivateKey) ed25519.PublicKey {
	return s.impl.ExtractPubKey(privateKey)
}

func (s *Service25519Openssl) Sign(privateKey ed25519.PrivateKey, message []byte) []byte {
	privKeyFilePath := filez.CreateTempFile("private_key.pem", []byte(s.PrivKeyExport(privateKey)))
	defer os.Remove(privKeyFilePath)

	msgFilePath := filez.CreateTempFile("message.txt", message)
	defer os.Remove(msgFilePath)

	sigFilePath := filez.CreateTempFile("signature.bin", nil)
	defer os.Remove(sigFilePath)

	err := execz.New(s.ctx, "openssl", "pkeyutl", "-sign", "-inkey", privKeyFilePath, "-rawin", "-in", msgFilePath, "-out", sigFilePath).Run()
	errorz.Check(err)

	signature := filez.MustReadFile(sigFilePath, 256)
	return signature.Bytes()
}

func (s *Service25519Openssl) Verify(publicKey ed25519.PublicKey, message []byte, signature []byte) bool {
	pubKeyFilePath := filez.CreateTempFile("public_key.pem", []byte(s.PubKeyExport(publicKey)))
	defer os.Remove(pubKeyFilePath)

	msgFilePath := filez.CreateTempFile("message.txt", message)
	defer os.Remove(msgFilePath)

	sigFilePath := filez.CreateTempFile("signature.bin", signature)
	defer os.Remove(sigFilePath)

	err := execz.New(s.ctx, "openssl", "pkeyutl", "-verify", "-pubin", "-inkey", pubKeyFilePath, "-rawin", "-in", msgFilePath, "-sigfile", sigFilePath).Run()
	return err == nil
}

func (s *Service25519Openssl) FindPubKey(pubs []ed25519.PublicKey, keyId string) ed25519.PublicKey {
	return s.impl.FindPubKey(pubs, keyId)
}

func (s *Service25519Openssl) PubEncrypt(publicKey ed25519.PublicKey, label []byte, message []byte) []byte {
	return s.impl.PubEncrypt(publicKey, label, message)
}

func (s *Service25519Openssl) PrivDecrypt(privateKey ed25519.PrivateKey, label []byte, ciphertext []byte) []byte {
	return s.impl.PrivDecrypt(privateKey, label, ciphertext)
}
