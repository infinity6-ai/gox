package cryptz25519

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
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

func TestUnitDeriveFromRSA(t *testing.T) {
	type testScenario struct {
		rsaKey   *rsa.PrivateKey
		errMsg   string
		validate func(t *testing.T, pub ed25519.PublicKey, priv ed25519.PrivateKey)
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		pub, priv, err := DeriveFromRSA(s.rsaKey)
		if s.errMsg != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.errMsg)
			require.Nil(t, pub)
			require.Nil(t, priv)
			return
		}
		require.NoError(t, err)
		require.NotNil(t, pub)
		require.NotNil(t, priv)
		if s.validate != nil {
			s.validate(t, pub, priv)
		}
	}

	rsaKey1, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	rsaKey2, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	t.Run("Nil RSA key returns error", func(t *testing.T) {
		check(t, testScenario{
			rsaKey: nil,
			errMsg: "rsaKey cannot be nil",
		})
	})

	t.Run("Valid RSA key derives valid Ed25519 keypair deterministically", func(t *testing.T) {
		var firstPub ed25519.PublicKey
		var firstPriv ed25519.PrivateKey

		check(t, testScenario{
			rsaKey: rsaKey1,
			validate: func(t *testing.T, pub ed25519.PublicKey, priv ed25519.PrivateKey) {
				require.Equal(t, ed25519.PublicKeySize, len(pub))
				require.Equal(t, ed25519.PrivateKeySize, len(priv))
				require.Equal(t, pub, priv.Public().(ed25519.PublicKey))

				// Sign and verify with derived keys to ensure cryptographic validity
				msg := []byte("deterministic derivation test message")
				sig := ed25519.Sign(priv, msg)
				require.True(t, ed25519.Verify(pub, msg, sig))

				firstPub = pub
				firstPriv = priv
			},
		})

		// Derive again from the same RSA key to test determinism
		check(t, testScenario{
			rsaKey: rsaKey1,
			validate: func(t *testing.T, pub ed25519.PublicKey, priv ed25519.PrivateKey) {
				require.Equal(t, firstPub, pub)
				require.Equal(t, firstPriv, priv)
			},
		})
	})

	t.Run("Different RSA keys derive different Ed25519 keys", func(t *testing.T) {
		pub1, priv1, err := DeriveFromRSA(rsaKey1)
		require.NoError(t, err)

		check(t, testScenario{
			rsaKey: rsaKey2,
			validate: func(t *testing.T, pub2 ed25519.PublicKey, priv2 ed25519.PrivateKey) {
				require.NotEqual(t, pub1, pub2)
				require.NotEqual(t, priv1, priv2)
			},
		})
	})

	t.Run("DeriveEd25519FromRSA alias produces identical output", func(t *testing.T) {
		pub1, priv1, err := DeriveFromRSA(rsaKey1)
		require.NoError(t, err)

		pub2, priv2, err := DeriveEd25519FromRSA(rsaKey1)
		require.NoError(t, err)

		require.Equal(t, pub1, pub2)
		require.Equal(t, priv1, priv2)
	})
}

