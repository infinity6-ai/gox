package cryptz25519

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitExtractPubKey(t *testing.T) {
	type testScenario struct {
		privKey    ed25519.PrivateKey
		wantPubKey ed25519.PublicKey
		panicMsg   string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.panicMsg != "" {
			require.PanicsWithValue(t, s.panicMsg, func() {
				ExtractPubKey(s.privKey)
			})
			return
		}
		got := ExtractPubKey(s.privKey)
		require.Equal(t, s.wantPubKey, got)
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	t.Run("Valid private key", func(t *testing.T) {
		check(t, testScenario{
			privKey:    priv,
			wantPubKey: pub,
		})
	})

	t.Run("Invalid private key length (empty)", func(t *testing.T) {
		check(t, testScenario{
			privKey:  ed25519.PrivateKey([]byte{}),
			panicMsg: "invalid ed25519 private key length",
		})
	})

	t.Run("Invalid private key length (32 bytes)", func(t *testing.T) {
		check(t, testScenario{
			privKey:  make(ed25519.PrivateKey, 32),
			panicMsg: "invalid ed25519 private key length",
		})
	})

	t.Run("Nil private key", func(t *testing.T) {
		check(t, testScenario{
			privKey:  nil,
			panicMsg: "invalid ed25519 private key length",
		})
	})
}
