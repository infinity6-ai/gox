package cryptzrsamsg

import (
	"crypto/rsa"
	"errors"
	"slices"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/cryptz/cryptzaes"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa"
	"go.code.infinity6.ai/platform/util/protoz"
)

type PublicKeyLoader func(id string) *rsa.PublicKey
type PrivateKeyLoader func(id string) *rsa.PrivateKey

type RSAMessagePlain struct {
	SrcKey string
	DstKey string
	Data   []byte
}

func (d *RSAMessagePlain) Validate() {
	checker.StrNotEmpty(d.SrcKey, "SrcKey")
	// validation.Hostname(d.SrcKey, "SrcKey")
	checker.StringRegex(`^[a-zA-z0-9\.\-_]+$`, d.SrcKey, "SrcKey")
	checker.StrNotEmpty(d.DstKey, "DstKey")
	// validation.Hostname(d.DstKey, "DstKey")
	checker.StringRegex(`^[a-zA-z0-9\.\-_]+$`, d.DstKey, "DstKey")
	checker.NotNil(d.Data, "Data")
}

type EncryptorKeysLoader struct {
	SrcKeyLoader PrivateKeyLoader
	DstKeyLoader PublicKeyLoader
}

func (k *EncryptorKeysLoader) Validate() {
	checker.NotNil(k.SrcKeyLoader, "SrcKeyLoader")
	checker.NotNil(k.DstKeyLoader, "DstKeyLoader")
}

type Encryptor struct {
	KeysLoader *EncryptorKeysLoader
	Message    *RSAMessagePlain
}

func (b *Encryptor) Validate() {
	checker.NotNil(b.KeysLoader, "KeysLoader")
	b.KeysLoader.Validate()
	checker.NotNil(b.Message, "Message")
	b.Message.Validate()
}

func dataToSign(msg *RSAMessageCiphered) []byte {
	return slices.Concat(msg.CipheredSymKey, []byte(msg.SrcKey), []byte(msg.DstKey), msg.CipheredData)
}

func (b *Encryptor) Encrypt() []byte {
	b.Validate()
	fromKey := b.KeysLoader.SrcKeyLoader(b.Message.SrcKey)
	toKey := b.KeysLoader.DstKeyLoader(b.Message.DstKey)
	s := cryptzrsa.NewService()
	symKey := cryptzaes.NewKey(32)
	cypheredSymKey := s.PubEncryptOEAP(toKey, nil, symKey)
	checker.Equal(toKey.Size(), len(cypheredSymKey), "cypheredSymKey length")
	cypheredMessage, err := cryptzaes.AESCrypt(symKey, b.Message.Data)
	errorz.Check(err)

	msg := &RSAMessageCiphered{}
	msg.Version = 1
	msg.SrcKey = b.Message.SrcKey
	msg.DstKey = b.Message.DstKey
	msg.CipheredSymKey = cypheredSymKey
	msg.CipheredData = cypheredMessage

	msg.Signature = s.Sign(fromKey, dataToSign(msg))
	checker.Equal(fromKey.Size(), len(msg.Signature), "sign length")

	ret := protoz.Format(msg)
	return ret
}

func NewEncryptor() *Encryptor {
	return &Encryptor{}
}

type DecryptorKeysLoader struct {
	SrcKeyLoader PublicKeyLoader
	DstKeyLoader PrivateKeyLoader
}

func (k *DecryptorKeysLoader) Validate() {
	checker.NotNil(k.SrcKeyLoader, "SrcKeyLoader")
	checker.NotNil(k.DstKeyLoader, "DstKeyLoader")
}

type Decryptor struct {
	KeysLoader   *DecryptorKeysLoader
	CipheredData []byte
}

func NewDecryptor() *Decryptor {
	return &Decryptor{}
}

func (d *Decryptor) Validate() {
	checker.NotNil(d.KeysLoader, "KeysLoader")
	d.KeysLoader.Validate()
	checker.NotNil(d.CipheredData, "CipheredData")
	checker.NotEmpty(d.CipheredData, "CipheredData")
}

func (d *Decryptor) Decrypt() *RSAMessagePlain {
	d.Validate()
	msg := protoz.Parse(d.CipheredData, &RSAMessageCiphered{})
	// validation.Hostname(string(msg.SrcKey), "SrcKey")
	// validation.Hostname(string(msg.DstKey), "DstKey")

	keyFrom := d.KeysLoader.SrcKeyLoader(string(msg.SrcKey))
	keyDest := d.KeysLoader.DstKeyLoader(string(msg.DstKey))
	checker.NotNil(keyFrom, "keyFrom")
	checker.NotNil(keyDest, "keyDest")

	msg.Validate(keyFrom.Size(), keyDest.Size())

	s := cryptzrsa.NewService()
	if !s.Verify(keyFrom, dataToSign(msg), msg.Signature) {
		panic(errors.New("sign fail"))
	}
	symKey := s.PrivDecryptOEAP(keyDest, nil, msg.CipheredSymKey)

	plainMessage, err := cryptzaes.AESDecrypt(symKey, msg.CipheredData)
	errorz.Check(err)

	ret := &RSAMessagePlain{}
	ret.SrcKey = string(msg.SrcKey)
	ret.DstKey = string(msg.DstKey)
	ret.Data = plainMessage.Bytes()
	return ret

}

func Send(payload *cryptzjwt.JWTPayload, keysLoader *EncryptorKeysLoader) []byte {
	checker.NotNil(payload, "payload")

	cryptzjwt.NewVerifier().
		AnyAudience().
		AnyIssuers().
		MustVerify(&cryptzjwt.JWTToken{Payload: payload})

	payloadData := protoz.Format(payload)

	encryptor := NewEncryptor()
	encryptor.KeysLoader = keysLoader
	encryptor.Message = &RSAMessagePlain{
		SrcKey: payload.Iss,
		DstKey: payload.Aud,
		Data:   payloadData,
	}

	ret := encryptor.Encrypt()
	return ret
}

func Receive(ciphertext []byte, keysLoader *DecryptorKeysLoader) *cryptzjwt.JWTPayload {
	decryptor := NewDecryptor()
	decryptor.KeysLoader = keysLoader
	decryptor.CipheredData = ciphertext

	plainMessage := decryptor.Decrypt()

	ret := protoz.Parse(plainMessage.Data, &cryptzjwt.JWTPayload{})

	cryptzjwt.NewVerifier().
		SetAudiences(plainMessage.DstKey).
		SetIssuers(plainMessage.SrcKey).
		MustVerify(&cryptzjwt.JWTToken{Payload: ret})

	return ret
}
