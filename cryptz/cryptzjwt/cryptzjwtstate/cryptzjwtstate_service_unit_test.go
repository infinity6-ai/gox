package cryptzjwtstate_test

import (
	"testing"
	"time"

	"github.com/infinity6-ai/gox/cryptz/cryptzaes"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt/cryptzjwtstate"
	"github.com/stretchr/testify/require"
)

func TestUnitStateEncryptDecrypt(t *testing.T) {
	key := cryptzaes.NewKey(32)

	payloadSent, err := cryptzjwt.NewToken().
		GenerateJti().
		SetPayloadType("m").
		CreatedRightNow().
		ExpiresIn(10 * time.Minute).
		SetIssuer("src").
		SetAudience("dst").
		SetSubject("mysub").
		SetExtension("").
		SetIdentityProvider("").
		BuildPayload()
	require.NoError(t, err)

	type testScenario struct {
		name string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		ciphertext := cryptzjwtstate.Send(payloadSent, &cryptzjwtstate.EncryptorKeysLoader{
			KeyLoader: func(srcKey, dstKey string) cryptzaes.Key {
				require.Equal(t, "src", srcKey)
				require.Equal(t, "dst", dstKey)
				return key
			},
		})
		require.NotNil(t, ciphertext)

		// Size check: symmetric payload should be very small (< 150 bytes)
		// compared to RSA's 376+ bytes
		require.Less(t, len(ciphertext), 150)

		payloadReceived := cryptzjwtstate.Receive(ciphertext, &cryptzjwtstate.DecryptorKeysLoader{
			KeyLoader: func(srcKey, dstKey string) cryptzaes.Key {
				require.Equal(t, "src", srcKey)
				require.Equal(t, "dst", dstKey)
				return key
			},
		})

		require.Equal(t, payloadSent, payloadReceived)
	}

	t.Run("Standard Send and Receive with KeyLoader", func(t *testing.T) {
		check(t, testScenario{name: "standard"})
	})
}

func TestUnitStateWithKeyHelpers(t *testing.T) {
	key := cryptzaes.NewKey(32)

	payloadSent, err := cryptzjwt.NewToken().
		GenerateJti().
		SetPayloadType("s").
		CreatedRightNow().
		ExpiresIn(5 * time.Minute).
		SetIssuer("myapp").
		SetAudience("myapp").
		SetSubject("user-123").
		SetExtension("").
		SetIdentityProvider("").
		BuildPayload()
	require.NoError(t, err)

	t.Run("SendWithKey and ReceiveWithKey binary roundtrip", func(t *testing.T) {
		ciphertext := cryptzjwtstate.SendWithKey(payloadSent, key)
		require.NotNil(t, ciphertext)
		require.Less(t, len(ciphertext), 150)

		payloadReceived := cryptzjwtstate.ReceiveWithKey(ciphertext, key)
		require.Equal(t, payloadSent, payloadReceived)
	})

	t.Run("SendStringWithKey and ReceiveStringWithKey URL-safe string roundtrip", func(t *testing.T) {
		tokenStr := cryptzjwtstate.SendStringWithKey(payloadSent, key)
		require.NotEmpty(t, tokenStr)
		// Compact base64 string (< 200 chars)
		require.Less(t, len(tokenStr), 200)

		payloadReceived := cryptzjwtstate.ReceiveStringWithKey(tokenStr, key)
		require.Equal(t, payloadSent, payloadReceived)
	})
}

func TestUnitStateTamperAndValidation(t *testing.T) {
	key := cryptzaes.NewKey(32)
	wrongKey := cryptzaes.NewKey(32)

	payload, err := cryptzjwt.NewToken().
		GenerateJti().
		SetPayloadType("m").
		CreatedRightNow().
		ExpiresIn(10 * time.Minute).
		SetIssuer("service-a").
		SetAudience("service-b").
		SetSubject("12345").
		SetExtension("").
		SetIdentityProvider("").
		BuildPayload()
	require.NoError(t, err)

	ciphertext := cryptzjwtstate.SendWithKey(payload, key)

	t.Run("Decryption with wrong key panics/fails", func(t *testing.T) {
		require.Panics(t, func() {
			cryptzjwtstate.ReceiveWithKey(ciphertext, wrongKey)
		})
	})

	t.Run("Tampered ciphertext bytes panics/fails", func(t *testing.T) {
		tampered := make([]byte, len(ciphertext))
		copy(tampered, ciphertext)
		// Mutate a byte in the encrypted data portion
		tampered[len(tampered)-5] ^= 0xFF

		require.Panics(t, func() {
			cryptzjwtstate.ReceiveWithKey(tampered, key)
		})
	})

	t.Run("Expired token fails verification", func(t *testing.T) {
		expiredPayload := &cryptzjwt.JWTPayload{
			Jti: "expired-jti",
			Typ: "m",
			Iat: time.Now().Add(-10 * time.Minute).Unix(),
			Exp: time.Now().Add(-1 * time.Minute).Unix(),
			Iss: "service-a",
			Aud: "service-b",
		}

		require.Panics(t, func() {
			cryptzjwtstate.SendWithKey(expiredPayload, key)
		})
	})
}
