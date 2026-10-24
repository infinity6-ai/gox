package cmdcryptz

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitKrAesCommand(t *testing.T) {
	tempDir := t.TempDir()
	krFile := filepath.Join(tempDir, "state_kr.json")

	t.Run("kr-aes init creates empty AES keyring", func(t *testing.T) {
		cmd := newKrAesInitCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{"-f", krFile})

		err := cmd.Execute()
		require.NoError(t, err)

		data, err := os.ReadFile(krFile)
		require.NoError(t, err)
		require.Equal(t, "{}", strings.TrimSpace(string(data)))
	})

	t.Run("kr-aes list on empty keyring prints Keyring is empty", func(t *testing.T) {
		cmd := newKrAesListCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{"-f", krFile})

		err := cmd.Execute()
		require.NoError(t, err)
		require.Equal(t, "Keyring is empty\n", outBuf.String())
	})

	t.Run("kr-aes gen generates owner key", func(t *testing.T) {
		cmd := newKrAesGenCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{"-f", krFile, "-o", "authservice", "-s", "32"})

		err := cmd.Execute()
		require.NoError(t, err)
		require.Contains(t, outBuf.String(), "Generated AES key for owner authservice")
	})

	t.Run("kr-aes gen-pair generates pairwise key", func(t *testing.T) {
		cmd := newKrAesGenPairCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{"-f", krFile, "--src", "web", "--dst", "api", "-s", "32"})

		err := cmd.Execute()
		require.NoError(t, err)
		require.Contains(t, outBuf.String(), "Generated pairwise AES key for web <-> api")
	})

	t.Run("kr-aes list shows both keys", func(t *testing.T) {
		cmd := newKrAesListCmd()
		var outBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetArgs([]string{"-f", krFile})

		err := cmd.Execute()
		require.NoError(t, err)
		require.Contains(t, outBuf.String(), "authservice")
		require.Contains(t, outBuf.String(), "api:web")
	})

	t.Run("kr-aes exp-key and exp-pair output hex keys", func(t *testing.T) {
		expKeyCmd := newKrAesExpKeyCmd()
		var keyBuf bytes.Buffer
		expKeyCmd.SetOut(&keyBuf)
		expKeyCmd.SetArgs([]string{"-f", krFile, "-o", "authservice"})

		err := expKeyCmd.Execute()
		require.NoError(t, err)
		require.Len(t, strings.TrimSpace(keyBuf.String()), 64)

		expPairCmd := newKrAesExpPairCmd()
		var pairBuf bytes.Buffer
		expPairCmd.SetOut(&pairBuf)
		expPairCmd.SetArgs([]string{"-f", krFile, "--src", "web", "--dst", "api"})

		err = expPairCmd.Execute()
		require.NoError(t, err)
		require.Len(t, strings.TrimSpace(pairBuf.String()), 64)
	})

	t.Run("kr-aes del removes owner key", func(t *testing.T) {
		delCmd := newKrAesDelCmd()
		var outBuf bytes.Buffer
		delCmd.SetOut(&outBuf)
		delCmd.SetArgs([]string{"-f", krFile, "-o", "authservice"})

		err := delCmd.Execute()
		require.NoError(t, err)

		listCmd := newKrAesListCmd()
		var listBuf bytes.Buffer
		listCmd.SetOut(&listBuf)
		listCmd.SetArgs([]string{"-f", krFile})
		err = listCmd.Execute()
		require.NoError(t, err)
		require.NotContains(t, listBuf.String(), "authservice")
		require.Contains(t, listBuf.String(), "api:web")
	})

	t.Run("kr-aes del-pair removes pairwise key", func(t *testing.T) {
		delPairCmd := newKrAesDelPairCmd()
		var outBuf bytes.Buffer
		delPairCmd.SetOut(&outBuf)
		delPairCmd.SetArgs([]string{"-f", krFile, "--src", "web", "--dst", "api"})

		err := delPairCmd.Execute()
		require.NoError(t, err)

		listCmd := newKrAesListCmd()
		var listBuf bytes.Buffer
		listCmd.SetOut(&listBuf)
		listCmd.SetArgs([]string{"-f", krFile})
		err = listCmd.Execute()
		require.NoError(t, err)
		require.Equal(t, "Keyring is empty\n", listBuf.String())
	})
}
