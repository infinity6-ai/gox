package cryptz25519service_test

import (
	"crypto/ed25519"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptz25519"
	"github.com/stretchr/testify/require"
)

func TestUnitSignVerify(t *testing.T) {
	s := cryptz25519.NewService()

	priv := s.PrivKeyCreate()
	pub := priv.Public().(ed25519.PublicKey)

	msg := []byte("hello world, signing this")
	sig := s.Sign(priv, msg)

	require.True(t, s.Verify(pub, msg, sig))

	// Corrupt signature
	sig[0] ^= 0xff
	require.False(t, s.Verify(pub, msg, sig))

	// Corrupt message
	sig[0] ^= 0xff // restore
	msg[0] ^= 0xff
	require.False(t, s.Verify(pub, msg, sig))
}

func TestUnitEncryptDecrypt(t *testing.T) {
	s := cryptz25519.NewService()

	priv := s.PrivKeyCreate()
	pub := priv.Public().(ed25519.PublicKey)

	msg := []byte("super secret message")
	label := []byte("test-label")

	crypted := s.PubEncrypt(pub, label, msg)
	require.GreaterOrEqual(t, len(crypted), 32) // ephemeral key size

	decrypted := s.PrivDecrypt(priv, label, crypted)
	require.Equal(t, msg, decrypted)

	// Wrong label
	require.Panics(t, func() {
		s.PrivDecrypt(priv, []byte("wrong-label"), crypted)
	})

	// Corrupt payload
	crypted[40] ^= 0xff
	require.Panics(t, func() {
		s.PrivDecrypt(priv, label, crypted)
	})
}

func TestUnitExportImport(t *testing.T) {
	s := cryptz25519.NewService()

	priv := s.PrivKeyCreate()
	pub := priv.Public().(ed25519.PublicKey)

	privPEM := s.PrivKeyExport(priv)
	pubPEM := s.PubKeyExport(pub)

	require.Contains(t, privPEM, "BEGIN PRIVATE KEY")
	require.Contains(t, pubPEM, "BEGIN PUBLIC KEY")

	importedPriv := s.PrivKeyImport(privPEM)
	importedPub := s.PubKeyImport(pubPEM)

	require.Equal(t, priv, importedPriv)
	require.Equal(t, pub, importedPub)
}

func TestUnitFindPubKeyId(t *testing.T) {
	s := cryptz25519.NewService()

	priv1 := s.PrivKeyCreate()
	pub1 := priv1.Public().(ed25519.PublicKey)

	priv2 := s.PrivKeyCreate()
	pub2 := priv2.Public().(ed25519.PublicKey)

	id1 := s.PubKeyId(pub1)
	id2 := s.PubKeyId(pub2)

	require.NotEmpty(t, id1)
	require.NotEmpty(t, id2)
	require.NotEqual(t, id1, id2)

	pubs := []ed25519.PublicKey{pub1, pub2}

	found1 := s.FindPubKey(pubs, id1)
	require.NotNil(t, found1)
	require.Equal(t, pub1, found1)

	foundMissing := s.FindPubKey(pubs, "NOTFOUND")
	require.Nil(t, foundMissing)
}
