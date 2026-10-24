package cmdcryptz

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzaes"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt/cryptzjwtstate"
	"github.com/stretchr/testify/require"
)

func TestUnitJwtStateCommand(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("jwtstate gen-key generates 32-byte hex key", func(t *testing.T) {
		cmd := newJwtStateGenKeyCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{"-s", "32"})

		err := cmd.Execute()
		require.NoError(t, err)
		keyHex := strings.TrimSpace(outBuf.String())
		require.Len(t, keyHex, 64)
	})

	t.Run("jwtstate encrypt and decrypt with direct key", func(t *testing.T) {
		key := cryptzaes.NewKey(32)
		keyHex := "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"
		_ = key

		encCmd := newJwtStateEncryptCmd()
		var encBuf bytes.Buffer
		encCmd.SetOut(&encBuf)
		encCmd.SetArgs([]string{
			"-k", keyHex,
			"--src", "session",
			"user_session_state_data",
		})
		err := encCmd.Execute()
		require.NoError(t, err)

		ciphertextHex := strings.TrimSpace(encBuf.String())

		decCmd := newJwtStateDecryptCmd()
		var decBuf bytes.Buffer
		decCmd.SetOut(&decBuf)
		decCmd.SetArgs([]string{
			"-k", keyHex,
			ciphertextHex,
		})
		err = decCmd.Execute()
		require.NoError(t, err)
		require.Equal(t, "user_session_state_data", decBuf.String())
	})

	t.Run("jwtstate encrypt and decrypt with state keyring", func(t *testing.T) {
		kr := cryptzjwtstate.NewKeyRing()
		kr.Generate("myservice", 32)
		krFile := filepath.Join(tempDir, "state_kr.json")
		krData, err := kr.Export()
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(krFile, krData, 0600))

		encCmd := newJwtStateEncryptCmd()
		var encBuf bytes.Buffer
		encCmd.SetOut(&encBuf)
		encCmd.SetArgs([]string{
			"--kr", krFile,
			"--src", "myservice",
			"oauth_state_12345",
		})
		err = encCmd.Execute()
		require.NoError(t, err)

		ciphertextHex := strings.TrimSpace(encBuf.String())

		decCmd := newJwtStateDecryptCmd()
		var decBuf bytes.Buffer
		decCmd.SetOut(&decBuf)
		decCmd.SetArgs([]string{
			"--kr", krFile,
			ciphertextHex,
		})
		err = decCmd.Execute()
		require.NoError(t, err)
		require.Equal(t, "oauth_state_12345", decBuf.String())
	})
}
