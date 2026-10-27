package cryptz25519service_test

import (
	"context"
	"crypto/ed25519"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519service"
	"github.com/stretchr/testify/require"
)

func checkKeyImport(t *testing.T, sCreateExport, sImportVerify cryptz25519service.Service25519) {
	// Create and export key using sCreateExport
	privOriginal := sCreateExport.PrivKeyCreate()
	pubOriginal := privOriginal.Public().(ed25519.PublicKey)
	exportedPriv := sCreateExport.PrivKeyExport(privOriginal)
	exportedPub := sCreateExport.PubKeyExport(pubOriginal)

	// Import key using sImportVerify
	privImported := sImportVerify.PrivKeyImport(exportedPriv)
	pubImported := sImportVerify.PubKeyImport(exportedPub)

	// Verify imported keys can be exported back (self-check for sImportVerify)
	require.Contains(t, sImportVerify.PrivKeyExport(privImported), "BEGIN PRIVATE KEY")
	require.Contains(t, sImportVerify.PubKeyExport(pubImported), "BEGIN PUBLIC KEY")

	// Sign with imported private key using sImportVerify
	sign := sImportVerify.Sign(privImported, []byte("mytext"))
	require.Len(t, sign, 64)

	// Verify with imported public key using sImportVerify
	require.True(t, sImportVerify.Verify(pubImported, []byte("mytext"), sign))
	require.False(t, sImportVerify.Verify(pubImported, []byte("any"), sign))
}

func TestUnitKeyImport(t *testing.T) {
	services := getImplementations(context.Background())
	for _, s1 := range services { // Service to create and export
		for _, s2 := range services { // Service to import and verify
			t.Run(buildName(s1, s2), func(t *testing.T) {
				checkKeyImport(t, s1, s2)
			})
		}
	}
}
