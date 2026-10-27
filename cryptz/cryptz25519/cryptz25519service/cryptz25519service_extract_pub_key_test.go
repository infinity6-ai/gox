package cryptz25519service_test

import (
	"context"
	"crypto/ed25519"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519service"
	"github.com/stretchr/testify/require"
)

func checkExtractPubKey(t *testing.T, service cryptz25519service.Service25519) {
	type testScenario struct {
		privKey    ed25519.PrivateKey
		wantPubKey ed25519.PublicKey
		panicMsg   string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.panicMsg != "" {
			require.PanicsWithValue(t, s.panicMsg, func() {
				service.ExtractPubKey(s.privKey)
			})
			return
		}
		got := service.ExtractPubKey(s.privKey)
		require.Equal(t, s.wantPubKey, got)
	}

	priv := service.PrivKeyCreate()
	pub := priv.Public().(ed25519.PublicKey)

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

func TestUnitExtractPubKey(t *testing.T) {
	services := getImplementations(context.Background())
	for _, s1 := range services {
		t.Run(buildName(s1), func(t *testing.T) {
			checkExtractPubKey(t, s1)
		})
	}
}
