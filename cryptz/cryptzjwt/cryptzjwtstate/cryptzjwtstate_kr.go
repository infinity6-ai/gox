package cryptzjwtstate

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/cryptz/cryptzaes"
)

var patternOwner = regexp.MustCompile(`^[a-z0-9\.\-_]+$`)

type KeyRing struct {
	keys map[string]cryptzaes.Key
}

func NewKeyRing() *KeyRing {
	return &KeyRing{
		keys: make(map[string]cryptzaes.Key),
	}
}

func pairKey(src, dst string) string {
	if src < dst {
		return src + ":" + dst
	}
	return dst + ":" + src
}

func (k *KeyRing) Add(owner string, key cryptzaes.Key) {
	checker.RegexMatch(patternOwner, owner, "owner")
	checker.NotEmpty(key, "key")
	k.keys[owner] = key
}

func (k *KeyRing) Generate(owner string, size int) cryptzaes.Key {
	key := cryptzaes.NewKey(size)
	k.Add(owner, key)
	return key
}

func (k *KeyRing) AddPair(src, dst string, key cryptzaes.Key) {
	checker.RegexMatch(patternOwner, src, "src")
	checker.RegexMatch(patternOwner, dst, "dst")
	checker.NotEmpty(key, "key")
	k.keys[pairKey(src, dst)] = key
}

func (k *KeyRing) GeneratePair(src, dst string, size int) cryptzaes.Key {
	key := cryptzaes.NewKey(size)
	k.AddPair(src, dst, key)
	return key
}

func (k *KeyRing) GetKey(src, dst string) cryptzaes.Key {
	// 1. Check exact symmetric pair (src, dst)
	if key, ok := k.keys[pairKey(src, dst)]; ok {
		return key
	}
	// 2. If src == dst, check self-owner key
	if src == dst {
		if key, ok := k.keys[src]; ok {
			return key
		}
	}
	// 3. Fallback to dst or src if registered as individual key
	if key, ok := k.keys[dst]; ok {
		return key
	}
	if key, ok := k.keys[src]; ok {
		return key
	}
	return nil
}

type MessageKeyRing struct {
	kr *KeyRing
}

func FromKeyRing(kr *KeyRing) *MessageKeyRing {
	checker.NotNil(kr, "kr")
	return &MessageKeyRing{kr: kr}
}

func (m *MessageKeyRing) KeyLoader() KeyLoader {
	return func(srcKey, dstKey string) cryptzaes.Key {
		return m.kr.GetKey(srcKey, dstKey)
	}
}

func (m *MessageKeyRing) EncryptorKeysLoader() *EncryptorKeysLoader {
	return &EncryptorKeysLoader{
		KeyLoader: m.KeyLoader(),
	}
}

func (m *MessageKeyRing) DecryptorKeysLoader() *DecryptorKeysLoader {
	return &DecryptorKeysLoader{
		KeyLoader: m.KeyLoader(),
	}
}

func (k *KeyRing) Export() ([]byte, error) {
	rec := make(map[string]string)
	for name, key := range k.keys {
		rec[name] = hex.EncodeToString(key)
	}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state keyring: %w", err)
	}
	return data, nil
}

func Import(data []byte) (*KeyRing, error) {
	kr := NewKeyRing()
	if err := kr.Import(data); err != nil {
		return nil, err
	}
	return kr, nil
}

func (k *KeyRing) Import(data []byte) error {
	var rec map[string]string
	if err := json.Unmarshal(data, &rec); err != nil {
		return fmt.Errorf("failed to parse state keyring JSON: %w", err)
	}
	if k.keys == nil {
		k.keys = make(map[string]cryptzaes.Key)
	}
	for name, hexKey := range rec {
		keyBytes, err := hex.DecodeString(hexKey)
		if err != nil {
			return fmt.Errorf("invalid hex key for %s: %w", name, err)
		}
		k.keys[name] = cryptzaes.Key(keyBytes)
	}
	return nil
}
