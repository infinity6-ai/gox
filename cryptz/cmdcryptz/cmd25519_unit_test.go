package cmdcryptz

import (
	"bytes"
	"crypto/ed25519"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptz25519"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa"
	"github.com/stretchr/testify/require"
)

func TestUnit25519Command(t *testing.T) {
	s := cryptz25519.NewService()
	priv := s.PrivKeyCreate()
	pub := priv.Public().(ed25519.PublicKey)
	privPEM := s.PrivKeyExport(priv)
	pubPEM := s.PubKeyExport(pub)
	expectedKeyId := s.PubKeyId(pub)

	tempDir := t.TempDir()
	privFile := filepath.Join(tempDir, "test_priv.pem")
	pubFile := filepath.Join(tempDir, "test_pub.pem")
	require.NoError(t, os.WriteFile(privFile, []byte(privPEM), 0600))
	require.NoError(t, os.WriteFile(pubFile, []byte(pubPEM), 0644))

	t.Run("25519 pub extracts public key correctly", func(t *testing.T) {
		cmd := new25519PubCmd()
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"-k", privFile})

		err := cmd.Execute()
		require.NoError(t, err)
		require.Contains(t, buf.String(), "-----BEGIN PUBLIC KEY-----")
		require.Equal(t, strings.TrimSpace(pubPEM), strings.TrimSpace(buf.String()))
	})

	t.Run("25519 id computes key ID", func(t *testing.T) {
		cmd := new25519IdCmd()
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"-k", pubFile})

		err := cmd.Execute()
		require.NoError(t, err)
		require.Equal(t, expectedKeyId, strings.TrimSpace(buf.String()))
	})

	t.Run("25519 sign and verify with hex", func(t *testing.T) {
		signCmd := new25519SignCmd()
		var signBuf bytes.Buffer
		var errBuf bytes.Buffer
		signCmd.SetOut(&signBuf)
		signCmd.SetErr(&errBuf)
		signCmd.SetArgs([]string{"-k", privFile, "message to sign"})

		err := signCmd.Execute()
		require.NoError(t, err)
		require.Empty(t, errBuf.String(), "25519 sign must not write output to stderr")
		sigHex := strings.TrimSpace(signBuf.String())
		require.NotEmpty(t, sigHex)

		// Verify valid signature
		verifyCmd := new25519VerifyCmd()
		var verifyBuf bytes.Buffer
		verifyCmd.SetOut(&verifyBuf)
		verifyCmd.SetArgs([]string{"-k", pubFile, "--sig", sigHex, "message to sign"})

		err = verifyCmd.Execute()
		require.NoError(t, err)
		require.Contains(t, verifyBuf.String(), "Verification succeeded")

		// Verify corrupted message fails
		badVerifyCmd := new25519VerifyCmd()
		var badVerifyBuf bytes.Buffer
		badVerifyCmd.SetOut(&badVerifyBuf)
		badVerifyCmd.SetArgs([]string{"-k", pubFile, "--sig", sigHex, "corrupted message"})
		err = badVerifyCmd.Execute()
		require.Error(t, err)
	})

	t.Run("25519 sign and verify with b64", func(t *testing.T) {
		signCmd := new25519SignCmd()
		var signBuf bytes.Buffer
		signCmd.SetOut(&signBuf)
		signCmd.SetArgs([]string{"-k", privFile, "--b64", "hello world"})

		err := signCmd.Execute()
		require.NoError(t, err)
		sigB64 := strings.TrimSpace(signBuf.String())

		verifyCmd := new25519VerifyCmd()
		var verifyBuf bytes.Buffer
		verifyCmd.SetOut(&verifyBuf)
		verifyCmd.SetArgs([]string{"-k", pubFile, "--sig", sigB64, "--b64", "hello world"})

		err = verifyCmd.Execute()
		require.NoError(t, err)
		require.Contains(t, verifyBuf.String(), "Verification succeeded")
	})

	t.Run("25519 encrypt and decrypt", func(t *testing.T) {
		encryptCmd := new25519EncryptCmd()
		var encBuf bytes.Buffer
		encryptCmd.SetOut(&encBuf)
		encryptCmd.SetArgs([]string{"-k", pubFile, "--label", "test-label", "confidential information"})

		err := encryptCmd.Execute()
		require.NoError(t, err)
		ciphertextHex := strings.TrimSpace(encBuf.String())

		decryptCmd := new25519DecryptCmd()
		var decBuf bytes.Buffer
		decryptCmd.SetOut(&decBuf)
		decryptCmd.SetArgs([]string{"-k", privFile, "--label", "test-label", ciphertextHex})

		err = decryptCmd.Execute()
		require.NoError(t, err)
		require.Equal(t, "confidential information", decBuf.String())
	})

	t.Run("25519 gen creates valid keys", func(t *testing.T) {
		genPriv := filepath.Join(tempDir, "gen_priv.pem")
		genPub := filepath.Join(tempDir, "gen_pub.pem")

		genCmd := new25519GenCmd()
		var genBuf bytes.Buffer
		genCmd.SetOut(&genBuf)
		genCmd.SetArgs([]string{"--priv-out", genPriv, "--pub-out", genPub})

		err := genCmd.Execute()
		require.NoError(t, err)

		privData, err := os.ReadFile(genPriv)
		require.NoError(t, err)
		require.Contains(t, string(privData), "-----BEGIN PRIVATE KEY-----")

		pubData, err := os.ReadFile(genPub)
		require.NoError(t, err)
		require.Contains(t, string(pubData), "-----BEGIN PUBLIC KEY-----")
	})

	t.Run("25519 derive-rsa derives keys deterministically", func(t *testing.T) {
		rsaService := cryptzrsa.NewService()
		rsaPriv := rsaService.PrivKeyCreate(2048)
		rsaPrivPEM := rsaService.PrivKeyExport(rsaPriv)
		rsaKeyFile := filepath.Join(tempDir, "rsa_priv.pem")
		require.NoError(t, os.WriteFile(rsaKeyFile, []byte(rsaPrivPEM), 0600))

		derivePriv := filepath.Join(tempDir, "derived_priv.pem")
		derivePub := filepath.Join(tempDir, "derived_pub.pem")

		deriveCmd := new25519DeriveRsaCmd()
		var deriveBuf bytes.Buffer
		deriveCmd.SetOut(&deriveBuf)
		deriveCmd.SetArgs([]string{"-k", rsaKeyFile, "--priv-out", derivePriv, "--pub-out", derivePub})

		err := deriveCmd.Execute()
		require.NoError(t, err)

		privBytes1, err := os.ReadFile(derivePriv)
		require.NoError(t, err)
		pubBytes1, err := os.ReadFile(derivePub)
		require.NoError(t, err)

		require.Contains(t, string(privBytes1), "-----BEGIN PRIVATE KEY-----")
		require.Contains(t, string(pubBytes1), "-----BEGIN PUBLIC KEY-----")

		// Re-derive to verify determinism
		derivePriv2 := filepath.Join(tempDir, "derived_priv2.pem")
		derivePub2 := filepath.Join(tempDir, "derived_pub2.pem")

		deriveCmd2 := new25519DeriveRsaCmd()
		var deriveBuf2 bytes.Buffer
		deriveCmd2.SetOut(&deriveBuf2)
		deriveCmd2.SetArgs([]string{"-k", rsaKeyFile, "--priv-out", derivePriv2, "--pub-out", derivePub2})

		err = deriveCmd2.Execute()
		require.NoError(t, err)

		privBytes2, err := os.ReadFile(derivePriv2)
		require.NoError(t, err)
		pubBytes2, err := os.ReadFile(derivePub2)
		require.NoError(t, err)

		require.Equal(t, string(privBytes1), string(privBytes2))
		require.Equal(t, string(pubBytes1), string(pubBytes2))

		// Verify that imported derived private key matches the derived public key
		importedPriv := s.PrivKeyImport(string(privBytes1))
		extractedPub := importedPriv.Public().(ed25519.PublicKey)
		require.Equal(t, strings.TrimSpace(string(pubBytes1)), strings.TrimSpace(s.PubKeyExport(extractedPub)))
	})

	t.Run("25519 derive-rsa fails with invalid RSA key", func(t *testing.T) {
		invalidKeyFile := filepath.Join(tempDir, "invalid.pem")
		require.NoError(t, os.WriteFile(invalidKeyFile, []byte("invalid pem data"), 0600))

		deriveCmd := new25519DeriveRsaCmd()
		var deriveBuf bytes.Buffer
		deriveCmd.SetOut(&deriveBuf)
		deriveCmd.SetArgs([]string{"-k", invalidKeyFile})

		err := deriveCmd.Execute()
		require.Error(t, err)
	})
}
