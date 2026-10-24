package cmdcryptz

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsakeyring"
	"github.com/stretchr/testify/require"
)

func TestUnitJwtMsgCommand(t *testing.T) {
	tempDir := t.TempDir()
	krFile := filepath.Join(tempDir, "keyring.json")

	// Setup keyring with alice and bob
	kr := cryptzrsakeyring.NewKeyRing()
	kr.Generate("alice", 1024)
	kr.Generate("bob", 1024)
	require.NoError(t, saveKeyRingFile(krFile, kr))

	var cipherHex string

	t.Run("jwtmsg encrypt produces ciphertext", func(t *testing.T) {
		cmd := newJwtMsgEncryptCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{
			"--kr", krFile,
			"--src", "alice",
			"--dst", "bob",
			"secret message from alice to bob",
		})

		err := cmd.Execute()
		require.NoError(t, err)

		cipherHex = strings.TrimSpace(outBuf.String())
		require.NotEmpty(t, cipherHex)
	})

	t.Run("jwtmsg decrypt recovers plaintext message", func(t *testing.T) {
		cmd := newJwtMsgDecryptCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{
			"--kr", krFile,
			cipherHex,
		})

		err := cmd.Execute()
		require.NoError(t, err)
		require.Equal(t, "secret message from alice to bob", outBuf.String())
	})

	t.Run("jwtmsg decrypt fails on tampered ciphertext", func(t *testing.T) {
		tampered := cipherHex[:len(cipherHex)-8] + "00000000"
		cmd := newJwtMsgDecryptCmd()
		var outBuf bytes.Buffer
		var errBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetErr(&errBuf)
		cmd.SetArgs([]string{
			"--kr", krFile,
			tampered,
		})

		err := cmd.Execute()
		require.Error(t, err)
	})
}
