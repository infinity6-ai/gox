package cryptzjwt_test

import (
	"crypto/rsa"
	"strings"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsakeyring"
	"github.com/stretchr/testify/require"
)

func TestUnitSignVerifyRSA(t *testing.T) {
	s := cryptzrsa.NewService()
	privKey := s.PrivKeyCreate(1024)
	pubKey := &privKey.PublicKey
	keyId := s.PubKeyId(pubKey)

	otherPriv := s.PrivKeyCreate(1024)
	otherPub := &otherPriv.PublicKey

	type testScenario struct {
		name      string
		claims    UserPayload
		tamper    func(token string) string
		verifyPub *rsa.PublicKey
		wantErr   string
	}

	check := func(t *testing.T, sc testScenario) {
		t.Helper()
		tokenStr, err := cryptzjwt.SignRSA(privKey, keyId, sc.claims)
		require.NoError(t, err)
		require.NotEmpty(t, tokenStr)

		if sc.tamper != nil {
			tokenStr = sc.tamper(tokenStr)
		}

		header, gotClaims, err := cryptzjwt.VerifyRSA[UserPayload](sc.verifyPub, tokenStr)
		if sc.wantErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), sc.wantErr)
			require.Nil(t, gotClaims)
			return
		}

		require.NoError(t, err)
		require.NotNil(t, header)
		require.Equal(t, "RS256", header.Alg)
		require.Equal(t, "JWT", header.Typ)
		require.Equal(t, keyId, header.KeyId())
		require.Equal(t, &sc.claims, gotClaims)
	}

	t.Run("Valid UserPayload succeeds with matching claims", func(t *testing.T) {
		check(t, testScenario{
			claims: UserPayload{
				StandardClaims: cryptzjwt.StandardClaims{
					Iss: "authservice",
					Sub: "user-rsa-42",
					Aud: "api",
					Exp: time.Now().Add(1 * time.Hour).Unix(),
					Typ: "usr",
				},
				Eml: "rsa@example.com",
			},
			verifyPub: pubKey,
		})
	})

	t.Run("Verification fails when using wrong public key", func(t *testing.T) {
		check(t, testScenario{
			claims: UserPayload{
				StandardClaims: cryptzjwt.StandardClaims{
					Iss: "authservice",
					Sub: "user-rsa-42",
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
					Sub: "user-rsa-42",
				},
			},
			verifyPub: pubKey,
			tamper: func(token string) string {
				parts := strings.Split(token, ".")
				tamperedSig := cryptzb64.UrlEncode([]byte("badrsasignaturethatdoesnotmatchthepayloadatall")).String()
				return parts[0] + "." + parts[1] + "." + tamperedSig
			},
			wantErr: "invalid signature",
		})
	})

	t.Run("Verification fails when token has invalid format", func(t *testing.T) {
		check(t, testScenario{
			claims: UserPayload{
				StandardClaims: cryptzjwt.StandardClaims{
					Sub: "user-rsa-42",
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

func TestUnitSealUnsealRSAKeyRing(t *testing.T) {
	kr := cryptzrsakeyring.NewKeyRing()
	kr.Generate("auth", 1024)

	otherKr := cryptzrsakeyring.NewKeyRing()
	otherKr.Generate("other", 1024)

	claims := UserPayload{
		StandardClaims: cryptzjwt.StandardClaims{
			Iss: "auth",
			Sub: "user-200",
			Aud: "service",
			Exp: time.Now().Add(30 * time.Minute).Unix(),
		},
		Eml: "bob@example.com",
	}

	header, tokenStr, err := cryptzjwt.SealRSA(kr, "auth", claims)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)
	require.Equal(t, "RS256", header.Alg)
	require.Equal(t, kr.Current("auth").KeyId(), header.KeyId())

	// Unseal with matching KeyRing
	unsealedHeader, gotClaims, err := cryptzjwt.UnsealRSA[UserPayload](kr, tokenStr)
	require.NoError(t, err)
	require.Equal(t, header, unsealedHeader)
	require.Equal(t, &claims, gotClaims)

	// Unseal with wrong KeyRing fails
	_, _, err = cryptzjwt.UnsealRSA[UserPayload](otherKr, tokenStr)
	require.Error(t, err)
	require.Contains(t, err.Error(), "key not found")
}
