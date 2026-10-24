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

	t.Run("keyring list on empty keyring prints Keyring is empty without panicking", func(t *testing.T) {
		cmd := newKeyringListCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{"-f", krFile})

		err := cmd.Execute()
		require.NoError(t, err)
		require.Equal(t, "Keyring is empty\n", outBuf.String())
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

	t.Run("keyring exp-priv and exp-pub export PEM keys for owner", func(t *testing.T) {
		// Test exp-priv for alice
		privCmd := newKeyringExpPrivCmd()
		var privBuf bytes.Buffer
		privCmd.SetOut(&privBuf)
		privCmd.SetArgs([]string{"-f", krFile, "-o", "alice"})

		err := privCmd.Execute()
		require.NoError(t, err)
		require.Contains(t, privBuf.String(), "-----BEGIN PRIVATE KEY-----")

		// Test exp-pub for alice
		pubCmd := newKeyringExpPubCmd()
		var pubBuf bytes.Buffer
		pubCmd.SetOut(&pubBuf)
		pubCmd.SetArgs([]string{"-f", krFile, "-o", "alice"})

		err = pubCmd.Execute()
		require.NoError(t, err)
		require.Contains(t, pubBuf.String(), "-----BEGIN PUBLIC KEY-----")

		// Test exp-pub to file
		pubFile := filepath.Join(tempDir, "alice_pub.pem")
		pubFileCmd := newKeyringExpPubCmd()
		var pubFileBuf bytes.Buffer
		pubFileCmd.SetOut(&pubFileBuf)
		pubFileCmd.SetArgs([]string{"-f", krFile, "-o", "alice", "--out", pubFile})

		err = pubFileCmd.Execute()
		require.NoError(t, err)

		pubData, err := os.ReadFile(pubFile)
		require.NoError(t, err)
		require.Contains(t, string(pubData), "-----BEGIN PUBLIC KEY-----")
	})

	t.Run("keyring del-priv removes only private key", func(t *testing.T) {
		delPrivCmd := newKeyringDelPrivCmd()
		var outBuf bytes.Buffer
		delPrivCmd.SetOut(&outBuf)
		delPrivCmd.SetArgs([]string{"-f", krFile, "-o", "bob"})

		err := delPrivCmd.Execute()
		require.NoError(t, err)
		require.Contains(t, outBuf.String(), "Deleted private key")

		// exp-priv for bob now fails
		privCmd := newKeyringExpPrivCmd()
		var privBuf bytes.Buffer
		privCmd.SetOut(&privBuf)
		privCmd.SetArgs([]string{"-f", krFile, "-o", "bob"})
		err = privCmd.Execute()
		require.Error(t, err)

		// exp-pub for bob still succeeds
		pubCmd := newKeyringExpPubCmd()
		var pubBuf bytes.Buffer
		pubCmd.SetOut(&pubBuf)
		pubCmd.SetArgs([]string{"-f", krFile, "-o", "bob"})
		err = pubCmd.Execute()
		require.NoError(t, err)
		require.Contains(t, pubBuf.String(), "-----BEGIN PUBLIC KEY-----")
	})

	t.Run("keyring del-key deletes entire key pair and cleans up owner", func(t *testing.T) {
		delKeyCmd := newKeyringDelKeyCmd()
		var outBuf bytes.Buffer
		delKeyCmd.SetOut(&outBuf)
		delKeyCmd.SetArgs([]string{"-f", krFile, "-o", "bob"})

		err := delKeyCmd.Execute()
		require.NoError(t, err)
		require.Contains(t, outBuf.String(), "Deleted key")

		// List should no longer have bob
		listCmd := newKeyringListCmd()
		var listBuf bytes.Buffer
		listCmd.SetOut(&listBuf)
		listCmd.SetArgs([]string{"-f", krFile})
		err = listCmd.Execute()
		require.NoError(t, err)
		require.NotContains(t, listBuf.String(), "bob")
		require.Contains(t, listBuf.String(), "alice")
	})
}
