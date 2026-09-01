package cryptzrsaopenssl

import (
	"context"
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/execz"
	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsaimpl"
)

type RSAServiceOpenssl struct {
	impl *cryptzrsaimpl.RSAServiceImpl
	ctx  context.Context
}

func New(ctx context.Context) *RSAServiceOpenssl {
	return &RSAServiceOpenssl{
		impl: cryptzrsaimpl.New(),
		ctx:  ctx,
	}
}

func (s *RSAServiceOpenssl) Name() string {
	return "cryptzrsaopenssl"
}

func (s *RSAServiceOpenssl) PrivKeyCreate(bits int) *rsa.PrivateKey {
	privKeyFilePath := filez.CreateTempFile("private_key.pem", nil) // Create empty temp file
	defer os.Remove(privKeyFilePath)

	err := execz.New(s.ctx, "openssl", "genrsa", "-out", privKeyFilePath, fmt.Sprintf("%d", bits)).Run()
	errorz.Check(err)

	keyData := filez.ReadFile(privKeyFilePath, 4096) // Max size for 2048-bit RSA key PEM
	return s.PrivKeyImport(keyData.String())
}

func (s *RSAServiceOpenssl) PrivKeyExport(privateKey *rsa.PrivateKey) string {
	return s.impl.PrivKeyExport(privateKey)
}

func (s *RSAServiceOpenssl) PubKeyExport(publicKey *rsa.PublicKey) string {
	return s.impl.PubKeyExport(publicKey)
}

func (s *RSAServiceOpenssl) PubKeyId(pub *rsa.PublicKey) string {
	return s.impl.PubKeyId(pub)
}

func (s *RSAServiceOpenssl) PrivKeyImport(key string) *rsa.PrivateKey {
	return s.impl.PrivKeyImport(key)
}

func (s *RSAServiceOpenssl) PubKeyImport(key string) *rsa.PublicKey {
	return s.impl.PubKeyImport(key)
}

func (s *RSAServiceOpenssl) Sign(privateKey *rsa.PrivateKey, message []byte) []byte {
	privKeyFilePath := filez.CreateTempFile("private_key.pem", []byte(s.PrivKeyExport(privateKey)))
	defer os.Remove(privKeyFilePath)

	msgFilePath := filez.CreateTempFile("message.txt", message)
	defer os.Remove(msgFilePath)

	sigFilePath := filez.CreateTempFile("signature.bin", nil) // Create empty temp file
	defer os.Remove(sigFilePath)

	err := execz.New(s.ctx, "openssl", "dgst", "-sha256", "-sign", privKeyFilePath, "-out", sigFilePath, msgFilePath).Run()
	errorz.Check(err)

	signature := filez.ReadFile(sigFilePath, 2048) // Max size for a 256-byte signature (RSA 2048-bit)
	return signature.Bytes()
}

func (s *RSAServiceOpenssl) Verify(publicKey *rsa.PublicKey, message []byte, signature []byte) bool {
	pubKeyFilePath := filez.CreateTempFile("public_key.pem", []byte(s.PubKeyExport(publicKey)))
	defer os.Remove(pubKeyFilePath)

	msgFilePath := filez.CreateTempFile("message.txt", message)
	defer os.Remove(msgFilePath)

	sigFilePath := filez.CreateTempFile("signature.bin", signature)
	defer os.Remove(sigFilePath)

	err := execz.New(s.ctx, "openssl", "dgst", "-sha256", "-verify", pubKeyFilePath, "-signature", sigFilePath, msgFilePath).Run()
	return err == nil
}

func (s *RSAServiceOpenssl) FindPubKey(pubs []*rsa.PublicKey, keyId string) *rsa.PublicKey {
	return s.impl.FindPubKey(pubs, keyId)
}

func (s *RSAServiceOpenssl) PubEncryptOEAP(publicKey *rsa.PublicKey, label []byte, message []byte) []byte {
	pubKeyFilePath := filez.CreateTempFile("public_key.pem", []byte(s.PubKeyExport(publicKey)))
	defer os.Remove(pubKeyFilePath)

	msgFilePath := filez.CreateTempFile("message.bin", message)
	defer os.Remove(msgFilePath)

	outFilePath := filez.CreateTempFile("out.bin", nil)
	defer os.Remove(outFilePath)

	args := []string{"pkeyutl", "-encrypt", "-pubin", "-inkey", pubKeyFilePath, "-in", msgFilePath, "-out", outFilePath, "-pkeyopt", "rsa_padding_mode:oaep", "-pkeyopt", "rsa_oaep_md:sha256"}
	if len(label) > 0 {
		// OpenSSL requires the label to be hex-encoded for the command line
		labelHex := fmt.Sprintf("%x", label)
		args = append(args, "-pkeyopt", "rsa_oaep_label:"+labelHex)
	}

	err := execz.New(s.ctx, "openssl", args...).Run()
	errorz.Check(err)

	return filez.ReadFile(outFilePath, 4096).Bytes() // Max size for RSA 2048-bit encrypted data
}

func (s *RSAServiceOpenssl) PrivDecryptOEAP(privateKey *rsa.PrivateKey, label []byte, ciphertext []byte) []byte {
	privKeyFilePath := filez.CreateTempFile("private_key.pem", []byte(s.PrivKeyExport(privateKey)))
	defer os.Remove(privKeyFilePath)

	cipherFilePath := filez.CreateTempFile("ciphertext.bin", ciphertext)
	defer os.Remove(cipherFilePath)

	outFilePath := filez.CreateTempFile("out.bin", nil)
	defer os.Remove(outFilePath)

	args := []string{"pkeyutl", "-decrypt", "-inkey", privKeyFilePath, "-in", cipherFilePath, "-out", outFilePath, "-pkeyopt", "rsa_padding_mode:oaep", "-pkeyopt", "rsa_oaep_md:sha256"}
	if len(label) > 0 {
		labelHex := fmt.Sprintf("%x", label)
		args = append(args, "-pkeyopt", "rsa_oaep_label:"+labelHex)
	}

	err := execz.New(s.ctx, "openssl", args...).Run()
	errorz.Check(err)

	return filez.ReadFile(outFilePath, 4096).Bytes()
}
