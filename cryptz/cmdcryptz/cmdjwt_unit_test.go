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

func TestUnitJwtCommand(t *testing.T) {
	s := cryptzrsa.NewService()
	priv := s.PrivKeyCreate(1024)
	privPEM := s.PrivKeyExport(priv)
	pubPEM := s.PubKeyExport(&priv.PublicKey)

	tempDir := t.TempDir()
	privFile := filepath.Join(tempDir, "jwt_priv.pem")
	pubFile := filepath.Join(tempDir, "jwt_pub.pem")
	require.NoError(t, os.WriteFile(privFile, []byte(privPEM), 0600))
	require.NoError(t, os.WriteFile(pubFile, []byte(pubPEM), 0644))

	var token string

	t.Run("jwt seal creates a valid token", func(t *testing.T) {
		cmd := newJwtSealCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{
			"-k", privFile,
			"--iss", "authservice",
			"--sub", "user123",
			"--aud", "api-gateway",
			"--exp", "2h",
		})

		err := cmd.Execute()
		require.NoError(t, err)

		token = strings.TrimSpace(outBuf.String())
		parts := strings.Split(token, ".")
		require.Len(t, parts, 3, "JWT must consist of 3 parts separated by dots")
	})

	t.Run("jwt inspect decodes token without key", func(t *testing.T) {
		cmd := newJwtInspectCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{token})

		err := cmd.Execute()
		require.NoError(t, err)

		output := outBuf.String()
		require.Contains(t, output, `"alg": "RS256"`)
		require.Contains(t, output, `"iss": "authservice"`)
		require.Contains(t, output, `"sub": "user123"`)
		require.Contains(t, output, `"aud": "api-gateway"`)
	})

	t.Run("jwt verify validates signature and assertions", func(t *testing.T) {
		cmd := newJwtVerifyCmd()
		var outBuf bytes.Buffer
		var errBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetErr(&errBuf)
		cmd.SetArgs([]string{
			"-k", pubFile,
			"--iss", "authservice",
			"--sub", "user123",
			"--aud", "api-gateway",
			token,
		})

		err := cmd.Execute()
		require.NoError(t, err)
		require.Empty(t, errBuf.String(), "jwt verify should not write errors on success")
		require.Contains(t, outBuf.String(), `"sub": "user123"`)
	})

	t.Run("jwt verify detects mismatched claims", func(t *testing.T) {
		cmd := newJwtVerifyCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{
			"-k", pubFile,
			"--sub", "wrong-user",
			token,
		})

		err := cmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "subject mismatch")
	})

	t.Run("jwt verify detects tampered token", func(t *testing.T) {
		tampered := token[:len(token)-5] + "AAAAA"
		cmd := newJwtVerifyCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{
			"-k", pubFile,
			tampered,
		})

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("jwt seal with custom iat and bxt-file", func(t *testing.T) {
		bxtFilePath := filepath.Join(tempDir, "bxt_source.bin")
		require.NoError(t, os.WriteFile(bxtFilePath, []byte{0x01, 0x02, 0x03, 0x04}, 0644))

		cmd := newJwtSealCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{
			"-k", privFile,
			"--iss", "authservice",
			"--sub", "device-008",
			"--aud", "api-gateway",
			"--iat", "-10m",
			"--bxt-file", bxtFilePath,
		})

		err := cmd.Execute()
		require.NoError(t, err)
		bxtToken := strings.TrimSpace(outBuf.String())

		// Verify with --bxt-out
		bxtExtracted := filepath.Join(tempDir, "bxt_extracted.bin")
		vCmd := newJwtVerifyCmd()
		var vOutBuf bytes.Buffer
		vCmd.SetOut(&vOutBuf)
		vCmd.SetArgs([]string{
			"-k", pubFile,
			"--bxt-out", bxtExtracted,
			bxtToken,
		})
		err = vCmd.Execute()
		require.NoError(t, err)

		extractedData, err := os.ReadFile(bxtExtracted)
		require.NoError(t, err)
		require.Equal(t, []byte{0x01, 0x02, 0x03, 0x04}, extractedData)
	})

	t.Run("jwt seal fails with conflicting bxt flags", func(t *testing.T) {
		cmd := newJwtSealCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{
			"-k", privFile,
			"--iss", "authservice",
			"--sub", "device-009",
			"--aud", "api-gateway",
			"--bxt-hex", "deadbeef",
			"--bxt-b64", "3q2+7w==",
		})

		err := cmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "none of the others can be")
	})

	t.Run("jwt inspect and unseal write bxt-out only on success", func(t *testing.T) {
		cmd := newJwtSealCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{
			"-k", privFile,
			"--iss", "authservice",
			"--sub", "device-file-test",
			"--aud", "api-gateway",
			"--bxt-hex", "cafebeef1234",
		})
		err := cmd.Execute()
		require.NoError(t, err)
		bxtToken := strings.TrimSpace(outBuf.String())

		// Test jwt inspect --bxt-out
		inspectOutPath := filepath.Join(tempDir, "inspect_bxt.bin")
		inspectCmd := newJwtInspectCmd()
		var insBuf bytes.Buffer
		inspectCmd.SetOut(&insBuf)
		inspectCmd.SetArgs([]string{
			"--bxt-out", inspectOutPath,
			bxtToken,
		})
		err = inspectCmd.Execute()
		require.NoError(t, err)

		extractedInspect, err := os.ReadFile(inspectOutPath)
		require.NoError(t, err)
		require.Equal(t, []byte{0xca, 0xfe, 0xbe, 0xef, 0x12, 0x34}, extractedInspect)

		// Test jwt verify --bxt-out succeeds
		verifyOutPath := filepath.Join(tempDir, "verify_bxt.bin")
		verifyCmd := newJwtVerifyCmd()
		var verBuf bytes.Buffer
		verifyCmd.SetOut(&verBuf)
		verifyCmd.SetArgs([]string{
			"-k", pubFile,
			"--bxt-out", verifyOutPath,
			bxtToken,
		})
		err = verifyCmd.Execute()
		require.NoError(t, err)

		extractedVerify, err := os.ReadFile(verifyOutPath)
		require.NoError(t, err)
		require.Equal(t, []byte{0xca, 0xfe, 0xbe, 0xef, 0x12, 0x34}, extractedVerify)

		// Test jwt verify --bxt-out does NOT write file if verification fails
		tamperedOutPath := filepath.Join(tempDir, "tampered_bxt.bin")
		tamperedToken := bxtToken[:len(bxtToken)-5] + "XXXXX"
		failVerifyCmd := newJwtVerifyCmd()
		var failBuf bytes.Buffer
		failVerifyCmd.SetOut(&failBuf)
		failVerifyCmd.SetArgs([]string{
			"-k", pubFile,
			"--bxt-out", tamperedOutPath,
			tamperedToken,
		})
		err = failVerifyCmd.Execute()
		require.Error(t, err)
		_, statErr := os.Stat(tamperedOutPath)
		require.True(t, os.IsNotExist(statErr), "bxt file must not be created if verification fails")
	})
}
