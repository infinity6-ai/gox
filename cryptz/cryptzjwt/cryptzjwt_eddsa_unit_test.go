package cryptzjwt_test

import (
	"crypto/ed25519"
	"strings"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/cryptz/cryptz25519"
	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519keyring"
	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
	"github.com/stretchr/testify/require"
)

type UserPayload struct {
	cryptzjwt.StandardClaims
	Eml string `json:"eml,omitempty"`
}

func TestUnitSignVerifyEdDSA(t *testing.T) {
	s := cryptz25519.NewService()
	privKey := s.PrivKeyCreate()
	pubKey := cryptz25519.ExtractPubKey(privKey)
	keyId := s.PubKeyId(pubKey)

	otherPriv := s.PrivKeyCreate()
	otherPub := cryptz25519.ExtractPubKey(otherPriv)

	type testScenario struct {
		name      string
		claims    UserPayload
		tamper    func(token string) string
		verifyPub ed25519.PublicKey
		wantErr   string
	}

	check := func(t *testing.T, sc testScenario) {
		t.Helper()
		tokenStr, err := cryptzjwt.SignEdDSA(privKey, keyId, sc.claims)
		require.NoError(t, err)
		require.NotEmpty(t, tokenStr)

		if sc.tamper != nil {
			tokenStr = sc.tamper(tokenStr)
		}

		header, gotClaims, err := cryptzjwt.VerifyEdDSA[UserPayload](sc.verifyPub, tokenStr)
		if sc.wantErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), sc.wantErr)
			require.Nil(t, gotClaims)
			return
		}

		require.NoError(t, err)
		require.NotNil(t, header)
		require.Equal(t, "EdDSA", header.Alg)
		require.Equal(t, "JWT", header.Typ)
		require.Equal(t, keyId, header.KeyId())
		require.Equal(t, &sc.claims, gotClaims)
	}

	t.Run("Valid UserPayload succeeds with matching claims", func(t *testing.T) {
		check(t, testScenario{
			claims: UserPayload{
				StandardClaims: cryptzjwt.StandardClaims{
					Iss: "authservice",
					Sub: "user-42",
					Aud: "api",
					Exp: time.Now().Add(1 * time.Hour).Unix(),
					Typ: "usr",
				},
				Eml: "test@example.com",
			},
			verifyPub: pubKey,
		})
	})

	t.Run("Verification fails when using wrong public key", func(t *testing.T) {
		check(t, testScenario{
			claims: UserPayload{
				StandardClaims: cryptzjwt.StandardClaims{
					Iss: "authservice",
					Sub: "user-42",
				},
			},
			verifyPub: otherPub,
			wantErr:   "invalid signature",
		})
	})

	t.Run("Verification fails when token signature is tampered", func(t *testing.T) {
		check(t, testScenario{
			claims: UserPayload{
				StandardClaims: cryptzjwt.StandardClaims{
					Iss: "authservice",
					Sub: "user-42",
				},
			},
			verifyPub: pubKey,
			tamper: func(token string) string {
				parts := strings.Split(token, ".")
				tamperedSig := cryptzb64.UrlEncode([]byte("badsignaturethatisinvalidandhasincorrectbytes")).String()
				return parts[0] + "." + parts[1] + "." + tamperedSig
			},
			wantErr: "invalid signature",
		})
	})

	t.Run("Verification fails when token has invalid format", func(t *testing.T) {
		check(t, testScenario{
			claims: UserPayload{
				StandardClaims: cryptzjwt.StandardClaims{
					Sub: "user-42",
				},
			},
			verifyPub: pubKey,
			tamper: func(token string) string {
				return "only.two.parts.extra"
			},
			wantErr: "invalid jwt format",
		})
	})
}

func TestUnitSealUnsealEdDSAKeyRing(t *testing.T) {
	kr := cryptz25519keyring.NewKeyRing()
	kr.Generate("auth")

	otherKr := cryptz25519keyring.NewKeyRing()
	otherKr.Generate("other")

	claims := UserPayload{
		StandardClaims: cryptzjwt.StandardClaims{
			Iss: "auth",
			Sub: "user-100",
			Aud: "service",
			Exp: time.Now().Add(30 * time.Minute).Unix(),
		},
		Eml: "alice@example.com",
	}

	header, tokenStr, err := cryptzjwt.SealEdDSA(kr, "auth", claims)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)
	require.Equal(t, "EdDSA", header.Alg)
	require.Equal(t, kr.Current("auth").KeyId(), header.KeyId())

	// Unseal with matching KeyRing
	unsealedHeader, gotClaims, err := cryptzjwt.UnsealEdDSA[UserPayload](kr, tokenStr)
	require.NoError(t, err)
	require.Equal(t, header, unsealedHeader)
	require.Equal(t, &claims, gotClaims)

	// Unseal with wrong KeyRing fails
	_, _, err = cryptzjwt.UnsealEdDSA[UserPayload](otherKr, tokenStr)
	require.Error(t, err)
	require.Contains(t, err.Error(), "key not found")
}
