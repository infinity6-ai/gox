package cmdcryptz

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzrsa"
	"github.com/stretchr/testify/require"
)

func TestUnitRsaCommand(t *testing.T) {
	s := cryptzrsa.NewService()
	priv := s.PrivKeyCreate(1024)
	privPEM := s.PrivKeyExport(priv)
	pubPEM := s.PubKeyExport(&priv.PublicKey)
	expectedKeyId := s.PubKeyId(&priv.PublicKey)

	tempDir := t.TempDir()
	privFile := filepath.Join(tempDir, "test_priv.pem")
	pubFile := filepath.Join(tempDir, "test_pub.pem")
	require.NoError(t, os.WriteFile(privFile, []byte(privPEM), 0600))
	require.NoError(t, os.WriteFile(pubFile, []byte(pubPEM), 0644))

	t.Run("rsa pub extracts public key correctly", func(t *testing.T) {
		cmd := newRsaPubCmd()
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"-k", privFile})

		err := cmd.Execute()
		require.NoError(t, err)
		require.Contains(t, buf.String(), "-----BEGIN PUBLIC KEY-----")
		require.Equal(t, strings.TrimSpace(pubPEM), strings.TrimSpace(buf.String()))
	})

	t.Run("rsa id computes key ID", func(t *testing.T) {
		cmd := newRsaIdCmd()
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"-k", pubFile})

		err := cmd.Execute()
		require.NoError(t, err)
		require.Equal(t, expectedKeyId, strings.TrimSpace(buf.String()))
	})

	t.Run("rsa sign and verify with hex", func(t *testing.T) {
		signCmd := newRsaSignCmd()
		var signBuf bytes.Buffer
		var errBuf bytes.Buffer
		signCmd.SetOut(&signBuf)
		signCmd.SetErr(&errBuf)
		signCmd.SetArgs([]string{"-k", privFile, "message to sign"})

		err := signCmd.Execute()
		require.NoError(t, err)
		require.Empty(t, errBuf.String(), "rsa sign must not write output to stderr")
		sigHex := strings.TrimSpace(signBuf.String())
		require.NotEmpty(t, sigHex)

		// Verify valid signature
		verifyCmd := newRsaVerifyCmd()
		var verifyBuf bytes.Buffer
		verifyCmd.SetOut(&verifyBuf)
		verifyCmd.SetArgs([]string{"-k", pubFile, "--sig", sigHex, "message to sign"})

		err = verifyCmd.Execute()
		require.NoError(t, err)
		require.Contains(t, verifyBuf.String(), "Verification succeeded")

		// Verify corrupted message fails
		badVerifyCmd := newRsaVerifyCmd()
		var badVerifyBuf bytes.Buffer
		badVerifyCmd.SetOut(&badVerifyBuf)
		badVerifyCmd.SetArgs([]string{"-k", pubFile, "--sig", sigHex, "corrupted message"})
		err = badVerifyCmd.Execute()
		require.Error(t, err)
	})

	t.Run("rsa sign and verify with b64", func(t *testing.T) {
		signCmd := newRsaSignCmd()
		var signBuf bytes.Buffer
		signCmd.SetOut(&signBuf)
		signCmd.SetArgs([]string{"-k", privFile, "--b64", "hello world"})

		err := signCmd.Execute()
		require.NoError(t, err)
		sigB64 := strings.TrimSpace(signBuf.String())

		verifyCmd := newRsaVerifyCmd()
		var verifyBuf bytes.Buffer
		verifyCmd.SetOut(&verifyBuf)
		verifyCmd.SetArgs([]string{"-k", pubFile, "--sig", sigB64, "--b64", "hello world"})

		err = verifyCmd.Execute()
		require.NoError(t, err)
		require.Contains(t, verifyBuf.String(), "Verification succeeded")
	})

	t.Run("rsa encrypt and decrypt OAEP", func(t *testing.T) {
		encryptCmd := newRsaEncryptCmd()
		var encBuf bytes.Buffer
		encryptCmd.SetOut(&encBuf)
		encryptCmd.SetArgs([]string{"-k", pubFile, "--label", "test-label", "confidential information"})

		err := encryptCmd.Execute()
		require.NoError(t, err)
		ciphertextHex := strings.TrimSpace(encBuf.String())

		decryptCmd := newRsaDecryptCmd()
		var decBuf bytes.Buffer
		decryptCmd.SetOut(&decBuf)
		decryptCmd.SetArgs([]string{"-k", privFile, "--label", "test-label", ciphertextHex})

		err = decryptCmd.Execute()
		require.NoError(t, err)
		require.Equal(t, "confidential information", decBuf.String())
	})

	t.Run("rsa gen creates valid keys", func(t *testing.T) {
		genPriv := filepath.Join(tempDir, "gen_priv.pem")
		genPub := filepath.Join(tempDir, "gen_pub.pem")

		genCmd := newRsaGenCmd()
		var genBuf bytes.Buffer
		genCmd.SetOut(&genBuf)
		genCmd.SetArgs([]string{"-b", "1024", "--priv-out", genPriv, "--pub-out", genPub})

		err := genCmd.Execute()
		require.NoError(t, err)

		privData, err := os.ReadFile(genPriv)
		require.NoError(t, err)
		require.Contains(t, string(privData), "-----BEGIN PRIVATE KEY-----")

		pubData, err := os.ReadFile(genPub)
		require.NoError(t, err)
		require.Contains(t, string(pubData), "-----BEGIN PUBLIC KEY-----")
	})
}
