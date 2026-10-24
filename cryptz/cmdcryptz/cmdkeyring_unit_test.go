package cmdcryptz

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitKeyringCommand(t *testing.T) {
	tempDir := t.TempDir()
	krFile := filepath.Join(tempDir, "keyring.json")

	t.Run("keyring init creates empty keyring file", func(t *testing.T) {
		cmd := newKeyringInitCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{"-f", krFile})

		err := cmd.Execute()
		require.NoError(t, err)

		data, err := os.ReadFile(krFile)
		require.NoError(t, err)
		require.Contains(t, string(data), `"owners"`)
	})

	t.Run("keyring init fails without required -f flag", func(t *testing.T) {
		cmd := newKeyringInitCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), `required flag(s) "file" not set`)
	})

	t.Run("keyring gen adds new owner and key", func(t *testing.T) {
		cmd := newKeyringGenCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{"-f", krFile, "-o", "alice", "-b", "1024"})

		err := cmd.Execute()
		require.NoError(t, err)
		require.Contains(t, outBuf.String(), "Generated key")

		// Add second owner
		outBuf.Reset()
		cmd.SetArgs([]string{"-f", krFile, "-o", "bob", "-b", "1024"})
		err = cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("keyring list displays owners", func(t *testing.T) {
		cmd := newKeyringListCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{"-f", krFile})

		err := cmd.Execute()
		require.NoError(t, err)
		require.Contains(t, outBuf.String(), "alice")
		require.Contains(t, outBuf.String(), "bob")
	})

	t.Run("keyring export outputs JSON", func(t *testing.T) {
		cmd := newKeyringExportCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{"-f", krFile})

		err := cmd.Execute()
		require.NoError(t, err)
		require.Contains(t, outBuf.String(), "-----BEGIN PRIVATE KEY-----")
		require.Contains(t, outBuf.String(), "-----BEGIN PUBLIC KEY-----")

		// Export pubs only
		outBuf.Reset()
		cmd.SetArgs([]string{"-f", krFile, "--pubs-only"})
		err = cmd.Execute()
		require.NoError(t, err)
		require.NotContains(t, outBuf.String(), "-----BEGIN PRIVATE KEY-----")
		require.Contains(t, outBuf.String(), "-----BEGIN PUBLIC KEY-----")
	})
}
