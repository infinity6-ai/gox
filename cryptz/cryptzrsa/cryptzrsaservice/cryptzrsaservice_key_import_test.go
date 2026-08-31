package cryptzrsaservice_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsaservice"
	"github.com/stretchr/testify/assert"
)

func checkKeyImport(t *testing.T, sCreateExport, sImportVerify cryptzrsaservice.RSAService) {
	// Create and export key using sCreateExport
	privOriginal := sCreateExport.PrivKeyCreate(2048)
	pubOriginal := &privOriginal.PublicKey
	exportedPriv := sCreateExport.PrivKeyExport(privOriginal)
	exportedPub := sCreateExport.PubKeyExport(pubOriginal)

	// Import key using sImportVerify
	privImported := sImportVerify.PrivKeyImport(exportedPriv)
	pubImported := sImportVerify.PubKeyImport(exportedPub)

	// Verify imported keys can be exported back (self-check for sImportVerify)
	assert.Contains(t, sImportVerify.PrivKeyExport(privImported), "BEGIN PRIVATE KEY")
	assert.Contains(t, sImportVerify.PubKeyExport(pubImported), "BEGIN PUBLIC KEY")

	// Sign with imported private key using sImportVerify
	sign := sImportVerify.Sign(privImported, []byte("mytext"))
	assert.Len(t, sign, 256)

	// Verify with imported public key using sImportVerify
	assert.True(t, sImportVerify.Verify(pubImported, []byte("mytext"), sign))
	assert.False(t, sImportVerify.Verify(pubImported, []byte("any"), sign))
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
