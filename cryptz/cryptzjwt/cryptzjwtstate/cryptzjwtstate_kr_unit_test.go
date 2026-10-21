package cryptzjwtstate_test

import (
	"testing"
	"time"

	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt/cryptzjwtstate"
	"github.com/stretchr/testify/require"
)

func TestUnitStateKeyRing(t *testing.T) {
	kr := cryptzjwtstate.NewKeyRing()
	kr.GeneratePair("src", "dst", 32)
	kr.Generate("myapp", 32)

	mkr := cryptzjwtstate.FromKeyRing(kr)

	type testScenario struct {
		name   string
		src    string
		dst    string
		sub    string
		expect bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		payloadSent, err := cryptzjwt.NewToken().
			GenerateJti().
			SetPayloadType("m").
			CreatedRightNow().
			ExpiresIn(10 * time.Minute).
			SetIssuer(s.src).
			SetAudience(s.dst).
			SetSubject(s.sub).
			SetExtension("").
			SetIdentityProvider("").
			BuildPayload()
		require.NoError(t, err)

		ciphertext := cryptzjwtstate.Send(payloadSent, mkr.EncryptorKeysLoader())
		require.NotNil(t, ciphertext)
		require.Less(t, len(ciphertext), 150)

		payloadReceived := cryptzjwtstate.Receive(ciphertext, mkr.DecryptorKeysLoader())
		require.Equal(t, payloadSent, payloadReceived)
	}

	t.Run("Sender to Receiver communication using pairwise key", func(t *testing.T) {
		check(t, testScenario{
			name:   "pairwise",
			src:    "src",
			dst:    "dst",
			sub:    "message-123",
			expect: true,
		})
	})

	t.Run("Self-state token using owner key", func(t *testing.T) {
		check(t, testScenario{
			name:   "self-state",
			src:    "myapp",
			dst:    "myapp",
			sub:    "oauth-state-abc",
			expect: true,
		})
	})
}

func TestUnitStateKeyRingUnknownKey(t *testing.T) {
	kr := cryptzjwtstate.NewKeyRing()
	kr.Generate("known", 32)
	mkr := cryptzjwtstate.FromKeyRing(kr)

	payload, err := cryptzjwt.NewToken().
		GenerateJti().
		SetPayloadType("m").
		CreatedRightNow().
		ExpiresIn(10 * time.Minute).
		SetIssuer("unknown-src").
		SetAudience("unknown-dst").
		SetSubject("sub").
		SetExtension("").
		SetIdentityProvider("").
		BuildPayload()
	require.NoError(t, err)

	t.Run("Fails when key is missing in keyring", func(t *testing.T) {
		require.Panics(t, func() {
			cryptzjwtstate.Send(payload, mkr.EncryptorKeysLoader())
		})
	})
}
