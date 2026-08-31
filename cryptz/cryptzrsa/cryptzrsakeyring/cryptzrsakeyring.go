package cryptzrsakeyring

import (
	"crypto/rsa"

	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa"
)

type Keys struct {
	keys map[string]*Key
}

func (k *Keys) Get(keyId string) *Key {
	key, ok := k.keys[keyId]
	if !ok {
		panic("key not found")
	}
	return key
}

type Key struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	keyId      string
	owner      string
}

func (k *Key) Validate() {
	if k.publicKey == nil {
		panic("publicKey is nil")
	}
	if k.privateKey == nil {
		panic("privateKey is nil")
	}
	checker.StrNotEmpty(k.keyId, "keyId")
	checker.StrNotEmpty(k.owner, "owner")
}

func (k *Key) Owner() string {
	return k.owner
}

func (k *Key) Public() *rsa.PublicKey {
	return k.publicKey
}

func (k *Key) Private() *rsa.PrivateKey {
	return k.privateKey
}

func (k *Key) KeyId() string {
	return k.keyId
}

type Owner struct {
	name    string
	keys    map[string]*Key
	current string
}

type KeyRing struct {
	owners map[string]*Owner
}

func (k *KeyRing) GetOwnerNames() []string {
	ret := make([]string, len(k.owners))
	i := 0
	for k := range k.owners {
		ret[i] = k
		i = i + 1
	}
	checker.NotEmpty(ret, "owners")
	return ret
}

func NewKeyRing() *KeyRing {
	return &KeyRing{}
}

func (k *KeyRing) Generate(owner string, bits int) {
	s := cryptzrsa.NewService()
	privKey := s.PrivKeyCreate(bits)
	k.AddPrivate(owner, privKey)
}

func (k *KeyRing) AddPrivate(owner string, privKey *rsa.PrivateKey) {
	key := k.AddPublic(owner, &privKey.PublicKey)
	key.privateKey = privKey
}

func (k *KeyRing) AddPublic(owner string, pubKey *rsa.PublicKey) *Key {
	checker.StringRegex("^[a-z][a-z0-9]*$", owner, "owner")
	s := cryptzrsa.NewService()
	keyId := s.PubKeyId(pubKey)
	if k.owners == nil {
		k.owners = make(map[string]*Owner)
	}
	o, ok := k.owners[owner]
	if !ok {
		o = &Owner{
			name: owner,
			keys: make(map[string]*Key),
		}
		k.owners[owner] = o
	}
	_, ok = o.keys[keyId]
	checker.False(ok, "key already exists: %s", keyId)
	key := &Key{
		owner:      owner,
		keyId:      keyId,
		publicKey:  pubKey,
		privateKey: nil,
	}
	o.keys[keyId] = key
	o.current = keyId
	return key
}

func (k *KeyRing) Current(owner string) *Key {
	o := k.getOwner(owner)
	ret := o.keys[o.current]
	if ret == nil {
		panic("not found")
	}
	return ret
}

func (k *KeyRing) getOwner(owner string) *Owner {
	if k.owners == nil {
		panic("not found")
	}
	o := k.owners[owner]
	if o == nil {
		panic("not found")
	}
	return o
}

func (k *KeyRing) getByKey(keyId string) *Key {
	for _, v := range k.owners {
		ret := v.keys[keyId]
		if ret != nil {
			return ret
		}
	}
	return nil
}

func (k *KeyRing) KeyIdOrNil(keyId string) *Key {
	ret := k.getByKey(keyId)
	if ret == nil {
		return nil
	}
	return ret
}

func (k *KeyRing) KeyId(keyId string) *Key {
	ret := k.getByKey(keyId)
	if ret == nil {
		panic("key not found")
	}
	return ret
}

func (k *KeyRing) Owners(owners ...string) *Keys {
	ret := &Keys{
		keys: make(map[string]*Key),
	}
	for _, owner := range owners {
		o := k.getOwner(owner)
		for keyId, key := range o.keys {
			ret.keys[keyId] = key
		}
	}
	if len(ret.keys) == 0 {
		panic("no keys")
	}
	return ret
}
