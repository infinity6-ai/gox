package cryptzaes

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/cryptz/cryptzrand"
	"go.code.infinity6.ai/platform/util"
	"go.code.infinity6.ai/platform/validation"
)

const MaxMessageLength = 10 * 1024 * 1024

type Key []byte

func NewKey(size int) Key {
	if size != 16 && size != 24 && size != 32 {
		util.Panic("invalid key size: must be 16, 24, or 32 bytes", nil)
	}
	return Key(cryptzrand.Rand(size))
}

func AESCrypt[T blobz.Data](key Key, plaindata T) []byte {
	block, err := aes.NewCipher(key)
	util.Check(err)
	gcm, err := cipher.NewGCM(block)
	util.Check(err)
	nonce := cryptzrand.Rand(gcm.NonceSize())
	plaintext := blobz.ToBytes(plaindata)
	validation.LesserEqual(len(plaintext), MaxMessageLength, "message length")
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext
}

func AESDecrypt(key Key, ciphertext []byte) (blobz.Blob, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext is too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return blobz.New(plaintext), nil
}
