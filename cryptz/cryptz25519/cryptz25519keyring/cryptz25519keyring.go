package cryptz25519keyring

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/cryptz/cryptz25519"
)

var patternOwner = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

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
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
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

func (k *Key) Public() ed25519.PublicKey {
	return k.publicKey
}

func (k *Key) Private() ed25519.PrivateKey {
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
	return ret
}

func NewKeyRing() *KeyRing {
	return &KeyRing{}
}

func (k *KeyRing) Generate(owner string) {
	s := cryptz25519.NewService()
	privKey := s.PrivKeyCreate()
	k.AddPrivate(owner, privKey)
}

func (k *KeyRing) AddPrivate(owner string, privKey ed25519.PrivateKey) {
	pubKey := privKey.Public().(ed25519.PublicKey)
	key := k.AddPublic(owner, pubKey)
	key.privateKey = privKey
}

func (k *KeyRing) AddPublic(owner string, pubKey ed25519.PublicKey) *Key {
	checker.RegexMatch(patternOwner, owner, "owner")
	s := cryptz25519.NewService()
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

func (k *KeyRing) DeleteKey(owner, keyId string) error {
	if k.owners == nil {
		return fmt.Errorf("owner %s not found", owner)
	}
	o, ok := k.owners[owner]
	if !ok {
		return fmt.Errorf("owner %s not found", owner)
	}
	if keyId == "" {
		keyId = o.current
	}
	if _, ok := o.keys[keyId]; !ok {
		return fmt.Errorf("key %s not found for owner %s", keyId, owner)
	}

	delete(o.keys, keyId)

	if len(o.keys) == 0 {
		delete(k.owners, owner)
		return nil
	}

	if o.current == keyId {
		for remainingId := range o.keys {
			o.current = remainingId
			break
		}
	}
	return nil
}

func (k *KeyRing) DeletePrivate(owner, keyId string) error {
	if k.owners == nil {
		return fmt.Errorf("owner %s not found", owner)
	}
	o, ok := k.owners[owner]
	if !ok {
		return fmt.Errorf("owner %s not found", owner)
	}
	if keyId == "" {
		keyId = o.current
	}
	key, ok := o.keys[keyId]
	if !ok {
		return fmt.Errorf("key %s not found for owner %s", keyId, owner)
	}
	if key.privateKey == nil {
		return fmt.Errorf("key %s for owner %s does not have a private key", keyId, owner)
	}
	key.privateKey = nil
	return nil
}

type keyRecord struct {
	KeyId      string `json:"key_id"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key,omitempty"`
}

type ownerRecord struct {
	Name    string                `json:"name"`
	Current string                `json:"current"`
	Keys    map[string]*keyRecord `json:"keys"`
}

type keyringRecord struct {
	Version int                     `json:"version"`
	Owners  map[string]*ownerRecord `json:"owners"`
}

func (k *KeyRing) ExportAll() ([]byte, error) {
	return k.export(false)
}

func (k *KeyRing) ExportPubs() ([]byte, error) {
	return k.export(true)
}

func (k *KeyRing) export(pubsOnly bool) ([]byte, error) {
	s := cryptz25519.NewService()
	rec := keyringRecord{
		Version: 1,
		Owners:  make(map[string]*ownerRecord),
	}

	for ownerName, owner := range k.owners {
		oRec := &ownerRecord{
			Name:    ownerName,
			Current: owner.current,
			Keys:    make(map[string]*keyRecord),
		}
		for keyId, key := range owner.keys {
			kRec := &keyRecord{
				KeyId:     keyId,
				PublicKey: s.PubKeyExport(key.publicKey),
			}
			if !pubsOnly && key.privateKey != nil {
				kRec.PrivateKey = s.PrivKeyExport(key.privateKey)
			}
			oRec.Keys[keyId] = kRec
		}
		rec.Owners[ownerName] = oRec
	}

	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal keyring: %w", err)
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

func (k *KeyRing) Import(data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to import keyring: %v", r)
		}
	}()

	var rec keyringRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return fmt.Errorf("failed to parse keyring JSON: %w", err)
	}

	s := cryptz25519.NewService()
	if k.owners == nil {
		k.owners = make(map[string]*Owner)
	}

	for ownerName, oRec := range rec.Owners {
		for _, kRec := range oRec.Keys {
			if kRec.PrivateKey != "" {
				priv := s.PrivKeyImport(kRec.PrivateKey)
				k.AddPrivate(ownerName, priv)
			} else if kRec.PublicKey != "" {
				pub := s.PubKeyImport(kRec.PublicKey)
				k.AddPublic(ownerName, pub)
			}
		}
		if owner, ok := k.owners[ownerName]; ok && oRec.Current != "" {
			owner.current = oRec.Current
		}
	}
	return nil
}
