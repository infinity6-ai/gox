package cryptzjwtmsg

import (
	"crypto/rsa"

	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsakeyring"
)

type MessageKyRing struct {
	kr *cryptzrsakeyring.KeyRing
}

func FromKeyRing(kr *cryptzrsakeyring.KeyRing) *MessageKyRing {
	return &MessageKyRing{kr: kr}
}

func (m *MessageKyRing) EncryptorKeysLoader() *EncryptorKeysLoader {
	return &EncryptorKeysLoader{
		SrcKeyLoader: func(id string) *rsa.PrivateKey {
			return m.kr.Current(id).Private()
		},
		DstKeyLoader: func(id string) *rsa.PublicKey {
			return m.kr.Current(id).Public()
		},
	}
}

func (m *MessageKyRing) DecryptorKeysLoader() *DecryptorKeysLoader {
	return &DecryptorKeysLoader{
		SrcKeyLoader: func(id string) *rsa.PublicKey {
			return m.kr.Current(id).Public()
		},
		DstKeyLoader: func(id string) *rsa.PrivateKey {
			return m.kr.Current(id).Private()
		},
	}
}
