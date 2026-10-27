package cryptz25519keyring_test

import (
	"crypto/ed25519"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptz25519"
	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519keyring"
	"github.com/stretchr/testify/require"
)

func TestUnitKeyRingGenerate(t *testing.T) {
	kr := cryptz25519keyring.NewKeyRing()
	ownerName := "testownergen"

	kr.Generate(ownerName)

	require.ElementsMatch(t, []string{ownerName}, kr.GetOwnerNames())

	currentKey := kr.Current(ownerName)
	require.NotNil(t, currentKey)
	require.NotNil(t, currentKey.Private())
	require.NotNil(t, currentKey.Public())
	require.Equal(t, ownerName, currentKey.Owner())
	require.NotEmpty(t, currentKey.KeyId())
}

func TestUnitKeyRingAddPublic(t *testing.T) {
	kr := cryptz25519keyring.NewKeyRing()
	ownerName := "testownerpub"

	s := cryptz25519.NewService()
	privKey := s.PrivKeyCreate()
	pubKey := privKey.Public().(ed25519.PublicKey)

	kr.AddPublic(ownerName, pubKey)

	currentKey := kr.Current(ownerName)
	require.NotNil(t, currentKey)
	require.Nil(t, currentKey.Private())
	require.NotNil(t, currentKey.Public())
}

func TestUnitKeyRingExportImport(t *testing.T) {
	kr := cryptz25519keyring.NewKeyRing()
	ownerName := "testowner"
	kr.Generate(ownerName)

	data, err := kr.ExportAll()
	require.NoError(t, err)

	kr2, err := cryptz25519keyring.Import(data)
	require.NoError(t, err)

	require.ElementsMatch(t, []string{ownerName}, kr2.GetOwnerNames())

	currentKey1 := kr.Current(ownerName)
	currentKey2 := kr2.Current(ownerName)

	require.Equal(t, currentKey1.KeyId(), currentKey2.KeyId())
	require.Equal(t, currentKey1.Private(), currentKey2.Private())
	require.Equal(t, currentKey1.Public(), currentKey2.Public())
}

func TestUnitKeyRingDelete(t *testing.T) {
	kr := cryptz25519keyring.NewKeyRing()
	ownerName := "testownerdel"
	kr.Generate(ownerName)

	keyId := kr.Current(ownerName).KeyId()

	err := kr.DeletePrivate(ownerName, keyId)
	require.NoError(t, err)
	require.Nil(t, kr.Current(ownerName).Private())
	require.NotNil(t, kr.Current(ownerName).Public())

	err = kr.DeleteKey(ownerName, keyId)
	require.NoError(t, err)

	require.Empty(t, kr.GetOwnerNames())
}
