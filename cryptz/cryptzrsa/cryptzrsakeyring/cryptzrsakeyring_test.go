package cryptzrsakeyring

import (
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.code.infinity6.ai/platform/cryptz/cryptzrsa"
)

// Helper function to create a new KeyRing for tests
func newTestKeyRing() *KeyRing {
	return &KeyRing{
		owners: make(map[string]*Owner),
	}
}

// Helper to generate a key pair
func generateKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	s := cryptzrsa.NewService()
	privKey := s.PrivKeyCreate(1024) // Use a smaller key size for faster tests
	assert.NotNil(t, privKey)
	assert.NotNil(t, privKey.PublicKey)
	return privKey, &privKey.PublicKey
}

func TestUnitGenerate(t *testing.T) {
	kr := newTestKeyRing()
	ownerName := "testownergen"
	bits := 1024

	kr.Generate(ownerName, bits)

	owner := kr.owners[ownerName]
	assert.NotNil(t, owner, "Owner should exist after Generate")
	assert.NotEmpty(t, owner.current, "Current key ID should be set")
	assert.Len(t, owner.keys, 1, "Should have one key")

	key := owner.keys[owner.current]
	assert.NotNil(t, key.privateKey, "Generated key should have a private key")
	assert.NotNil(t, key.publicKey, "Generated key should have a public key")
}

func TestUnitAddPublic(t *testing.T) {
	kr := newTestKeyRing()
	ownerName := "testownerpuba"
	_, pubKey := generateKeyPair(t)

	addedKey := kr.AddPublic(ownerName, pubKey)

	owner := kr.owners[ownerName]
	assert.NotNil(t, owner, "Owner should exist after AddPublic")
	assert.NotEmpty(t, owner.current, "Current key ID should be set")
	assert.Len(t, owner.keys, 1, "Should have one key")

	key := owner.keys[owner.current]
	assert.NotNil(t, key.publicKey, "Added key should have a public key")
	assert.Nil(t, key.privateKey, "Added key should not have a private key")
	assert.Equal(t, pubKey, addedKey.publicKey, "Returned key should match")

	// Test adding a second public key for the same owner
	_, pubKey2 := generateKeyPair(t)
	addedKey2 := kr.AddPublic(ownerName, pubKey2)
	assert.Len(t, owner.keys, 2, "Should have two keys after adding another")
	assert.Equal(t, pubKey2, addedKey2.publicKey, "Second returned key should match")
	assert.Equal(t, addedKey2.publicKey, kr.Current(ownerName).publicKey, "Current key should be the last added")
}

func TestUnitAddPublicDuplicateKey(t *testing.T) {
	kr := newTestKeyRing()
	ownerName := "testownerdupa"
	_, pubKey := generateKeyPair(t)

	kr.AddPublic(ownerName, pubKey)

	// Attempt to add the same key again
	assert.Panics(t, func() {
		kr.AddPublic(ownerName, pubKey)
	}, "Adding duplicate public key should panic")
}

func TestUnitAddPublicInvalidOwnerName(t *testing.T) {
	kr := newTestKeyRing()
	_, pubKey := generateKeyPair(t)

	assert.Panics(t, func() {
		kr.AddPublic("Invalid-Owner", pubKey)
	}, "Adding with invalid owner name should panic")

	assert.Panics(t, func() {
		kr.AddPublic("", pubKey)
	}, "Adding with empty owner name should panic")
}

func TestUnitAddPrivate(t *testing.T) {
	kr := newTestKeyRing()
	ownerName := "testownerpriva"
	privKey, _ := generateKeyPair(t)

	kr.AddPrivate(ownerName, privKey)

	owner := kr.owners[ownerName]
	assert.NotNil(t, owner, "Owner should exist after AddPrivate")
	assert.NotEmpty(t, owner.current, "Current key ID should be set")
	assert.Len(t, owner.keys, 1, "Should have one key")

	key := owner.keys[owner.current]
	assert.NotNil(t, key.privateKey, "Added key should have a private key")
	assert.NotNil(t, key.publicKey, "Added key should have a public key")
	assert.Equal(t, privKey, key.privateKey, "Private key should match")
	assert.Equal(t, &privKey.PublicKey, key.publicKey, "Public key should match")
}

func TestUnitCurrent(t *testing.T) {
	kr := newTestKeyRing()
	ownerName := "testownercurrenta"
	privKey, pubKey := generateKeyPair(t)

	kr.AddPrivate(ownerName, privKey)
	currentKey := kr.Current(ownerName)

	assert.NotNil(t, currentKey)
	assert.Equal(t, privKey, currentKey.privateKey)
	assert.Equal(t, pubKey, currentKey.publicKey)

	// Test with a second key
	privKey2, pubKey2 := generateKeyPair(t)
	kr.AddPrivate(ownerName, privKey2)
	currentKey = kr.Current(ownerName)
	assert.Equal(t, pubKey2, currentKey.publicKey, "Current key should be the last private key added")

	// Test Current on an owner with only a public key
	ownerNamePub := "testownercurrentpub"
	_, pubKey3 := generateKeyPair(t)
	kr.AddPublic(ownerNamePub, pubKey3)
	currentKeyPub := kr.Current(ownerNamePub)
	assert.NotNil(t, currentKeyPub)
	assert.Equal(t, pubKey3, currentKeyPub.publicKey)
	assert.Nil(t, currentKeyPub.privateKey)
}

func TestUnitCurrentNoOwner(t *testing.T) {
	kr := newTestKeyRing()
	assert.Panics(t, func() {
		kr.Current("nonexistent")
	}, "Current for nonexistent owner should panic")
}

func TestUnitKeyId(t *testing.T) {
	kr := newTestKeyRing()
	ownerName1 := "ownerid1"
	ownerName2 := "ownerid2"
	privKey1, pubKey1 := generateKeyPair(t)
	privKey2, pubKey2 := generateKeyPair(t)

	s := cryptzrsa.NewService()
	keyId1 := s.PubKeyId(pubKey1)
	keyId2 := s.PubKeyId(pubKey2)

	kr.AddPrivate(ownerName1, privKey1)
	kr.AddPrivate(ownerName2, privKey2)

	foundKey1 := kr.KeyId(keyId1)
	assert.NotNil(t, foundKey1)
	assert.Equal(t, privKey1, foundKey1.privateKey)

	foundKey2 := kr.KeyId(keyId2)
	assert.NotNil(t, foundKey2)
	assert.Equal(t, privKey2, foundKey2.privateKey)
}

func TestUnitKeyIdNotFound(t *testing.T) {
	kr := newTestKeyRing()
	assert.Panics(t, func() {
		kr.KeyId("nonexistentkeyid")
	}, "KeyId for nonexistent key should panic")
}

func TestUnitOwners(t *testing.T) {
	kr := newTestKeyRing()
	ownerName1 := "ownera"
	ownerName2 := "ownerb"

	privKeyA1, pubKeyA1 := generateKeyPair(t)
	privKeyA2, pubKeyA2 := generateKeyPair(t)
	privKeyB1, pubKeyB1 := generateKeyPair(t)

	kr.AddPrivate(ownerName1, privKeyA1)
	kr.AddPrivate(ownerName1, privKeyA2)
	kr.AddPrivate(ownerName2, privKeyB1)

	// Test for single owner
	keysA := kr.Owners(ownerName1)
	assert.NotNil(t, keysA)
	assert.Len(t, keysA.keys, 2)
	s := cryptzrsa.NewService()
	keyIdA1 := s.PubKeyId(pubKeyA1)
	keyIdA2 := s.PubKeyId(pubKeyA2)

	// Check if keys are in the map
	key1Found := keysA.Get(keyIdA1)
	assert.NotNil(t, key1Found)
	key2Found := keysA.Get(keyIdA2)
	assert.NotNil(t, key2Found)
	assert.Equal(t, pubKeyA1, key1Found.publicKey)
	assert.Equal(t, pubKeyA2, key2Found.publicKey)

	// Test for multiple owners
	keysAB := kr.Owners(ownerName1, ownerName2)
	assert.NotNil(t, keysAB)
	assert.Len(t, keysAB.keys, 3) // 2 from A, 1 from B
	keyIdB1 := s.PubKeyId(pubKeyB1)
	assert.NotNil(t, keysAB.Get(keyIdA1))
	assert.NotNil(t, keysAB.Get(keyIdA2))
	assert.NotNil(t, keysAB.Get(keyIdB1))

	// Test for an owner with only a public key
	ownerNameC := "ownerc"
	_, pubKeyC1 := generateKeyPair(t)
	kr.AddPublic(ownerNameC, pubKeyC1)
	keysC := kr.Owners(ownerNameC)
	assert.Len(t, keysC.keys, 1)
	keyIdC1 := s.PubKeyId(pubKeyC1)
	keyCFound := keysC.Get(keyIdC1)
	assert.NotNil(t, keyCFound)
	assert.Equal(t, pubKeyC1, keyCFound.publicKey)
	assert.Nil(t, keyCFound.privateKey)
}

func TestUnitOwnersNoKeysFound(t *testing.T) {
	kr := newTestKeyRing()
	assert.Panics(t, func() {
		kr.Owners("nonexistentowner")
	}, "Owners with no keys found should panic")

	// Add an owner but no keys
	// NOTE: This scenario is implicitly handled by `getOwner` which will panic if owner not found
	// or if owner has no keys, then `len(ret.keys) == 0` will cause panic later.
	// The `getOwner` helper within `KeyRing` will panic if `k.owners` is nil or the owner is not found.
	// If `k.owners` is not nil but the owner has no keys, it returns an empty `ret.keys` which then panics.
}