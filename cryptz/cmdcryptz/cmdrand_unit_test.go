package cmdcryptz

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
	"github.com/stretchr/testify/require"
)

func TestUnitRandCommand(t *testing.T) {
	type testScenario struct {
		args        []string
		validate    func(t *testing.T, output []byte)
		expectedErr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		cmd := newRandCmd()
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs(s.args)

		err := cmd.Execute()
		if s.expectedErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedErr)
			return
		}

		require.NoError(t, err)
		if s.validate != nil {
			s.validate(t, buf.Bytes())
		}
	}

	t.Run("Default output is hex of 32 bytes", func(t *testing.T) {
		check(t, testScenario{
			args: []string{},
			validate: func(t *testing.T, output []byte) {
				text := strings.TrimSpace(string(output))
				decoded, err := hex.DecodeString(text)
				require.NoError(t, err)
				require.Len(t, decoded, 32)
			},
		})
	})

	t.Run("Custom size hex", func(t *testing.T) {
		check(t, testScenario{
			args: []string{"-s", "16"},
			validate: func(t *testing.T, output []byte) {
				text := strings.TrimSpace(string(output))
				decoded, err := hex.DecodeString(text)
				require.NoError(t, err)
				require.Len(t, decoded, 16)
			},
		})
	})

	t.Run("Base64 output", func(t *testing.T) {
		check(t, testScenario{
			args: []string{"-s", "24", "--b64"},
			validate: func(t *testing.T, output []byte) {
				text := strings.TrimSpace(string(output))
				decoded, err := cryptzb64.StdDecode(text)
				require.NoError(t, err)
				require.Len(t, decoded.Bytes(), 24)
			},
		})
	})

	t.Run("Base64URL output", func(t *testing.T) {
		check(t, testScenario{
			args: []string{"-s", "48", "--b64url"},
			validate: func(t *testing.T, output []byte) {
				text := strings.TrimSpace(string(output))
				decoded, err := cryptzb64.UrlDecode(text)
				require.NoError(t, err)
				require.Len(t, decoded.Bytes(), 48)
				require.NotContains(t, text, "+")
				require.NotContains(t, text, "/")
			},
		})
	})

	t.Run("Raw binary output", func(t *testing.T) {
		check(t, testScenario{
			args: []string{"-s", "64", "--raw"},
			validate: func(t *testing.T, output []byte) {
				require.Len(t, output, 64)
			},
		})
	})

	t.Run("Mutually exclusive flags fail", func(t *testing.T) {
		check(t, testScenario{
			args:        []string{"--b64", "--raw"},
			expectedErr: "none of the others can be",
		})
	})
}
