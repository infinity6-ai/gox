package cryptzjwtstate

import (
	"errors"
	"regexp"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/protoz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/cryptz/cryptzaes"
	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
)

var patternKey = regexp.MustCompile(`^[a-zA-z0-9\.\-_]+$`)

type KeyLoader func(srcKey string, dstKey string) cryptzaes.Key

type StateMessagePlain struct {
	SrcKey string
	DstKey string
	KeyId  string
	Data   []byte
}

func (d *StateMessagePlain) Validate() {
	checker.StrNotEmpty(d.SrcKey, "SrcKey")
	checker.RegexMatch(patternKey, d.SrcKey, "SrcKey")
	checker.StrNotEmpty(d.DstKey, "DstKey")
	checker.RegexMatch(patternKey, d.DstKey, "DstKey")
	checker.NotNil(d.Data, "Data")
}

type EncryptorKeysLoader struct {
	KeyLoader KeyLoader
}

func (k *EncryptorKeysLoader) Validate() {
	checker.NotNil(k.KeyLoader, "KeyLoader")
}

type Encryptor struct {
	KeysLoader *EncryptorKeysLoader
	Message    *StateMessagePlain
}

func NewEncryptor() *Encryptor {
	return &Encryptor{}
}

func (b *Encryptor) Validate() {
	checker.NotNil(b.KeysLoader, "KeysLoader")
	b.KeysLoader.Validate()
	checker.NotNil(b.Message, "Message")
	b.Message.Validate()
}

func (b *Encryptor) Encrypt() []byte {
	b.Validate()
	key := b.KeysLoader.KeyLoader(b.Message.SrcKey, b.Message.DstKey)
	if len(key) == 0 {
		panic(errors.New("encryption key not found"))
	}

	cipheredMessage, err := cryptzaes.AESCrypt(key, b.Message.Data)
	errorz.Check(err)

	msg := &StateMessageCiphered{
		Version:      1,
		SrcKey:       b.Message.SrcKey,
		DstKey:       b.Message.DstKey,
		KeyId:        b.Message.KeyId,
		CipheredData: cipheredMessage,
	}

	ret := protoz.Format(msg)
	return ret
}

type DecryptorKeysLoader struct {
	KeyLoader KeyLoader
}

func (k *DecryptorKeysLoader) Validate() {
	checker.NotNil(k.KeyLoader, "KeyLoader")
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

func (d *Decryptor) Decrypt() *StateMessagePlain {
	d.Validate()
	msg := protoz.Parse(d.CipheredData, &StateMessageCiphered{})
	msg.Validate()

	key := d.KeysLoader.KeyLoader(msg.SrcKey, msg.DstKey)
	if len(key) == 0 {
		panic(errors.New("decryption key not found"))
	}

	plainMessage, err := cryptzaes.AESDecrypt(key, msg.CipheredData)
	errorz.Check(err)

	ret := &StateMessagePlain{
		SrcKey: msg.SrcKey,
		DstKey: msg.DstKey,
		KeyId:  msg.KeyId,
		Data:   plainMessage.Bytes(),
	}
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
	encryptor.Message = &StateMessagePlain{
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

func SingleKeyLoader(key cryptzaes.Key) KeyLoader {
	return func(srcKey, dstKey string) cryptzaes.Key {
		return key
	}
}

func SendWithKey(payload *cryptzjwt.JWTPayload, key cryptzaes.Key) []byte {
	return Send(payload, &EncryptorKeysLoader{KeyLoader: SingleKeyLoader(key)})
}

func ReceiveWithKey(ciphertext []byte, key cryptzaes.Key) *cryptzjwt.JWTPayload {
	return Receive(ciphertext, &DecryptorKeysLoader{KeyLoader: SingleKeyLoader(key)})
}

func SendString(payload *cryptzjwt.JWTPayload, keysLoader *EncryptorKeysLoader) string {
	bytes := Send(payload, keysLoader)
	return cryptzb64.UrlEncode(bytes).String()
}

func ReceiveString(token string, keysLoader *DecryptorKeysLoader) *cryptzjwt.JWTPayload {
	decoded, err := cryptzb64.UrlDecode(token)
	errorz.Check(err)
	return Receive(decoded.Bytes(), keysLoader)
}

func SendStringWithKey(payload *cryptzjwt.JWTPayload, key cryptzaes.Key) string {
	return SendString(payload, &EncryptorKeysLoader{KeyLoader: SingleKeyLoader(key)})
}

func ReceiveStringWithKey(token string, key cryptzaes.Key) *cryptzjwt.JWTPayload {
	return ReceiveString(token, &DecryptorKeysLoader{KeyLoader: SingleKeyLoader(key)})
}
