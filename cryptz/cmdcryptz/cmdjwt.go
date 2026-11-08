package cmdcryptz

import (
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/infinity6-ai/gox/cryptz/cryptz25519"
	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519keyring"
	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsakeyring"
	"github.com/spf13/cobra"
)

func newJwtCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jwt",
		Short: "JWT utilities (inspect, seal, verify)",
	}

	cmd.AddCommand(
		newJwtInspectCmd(),
		newJwtSealCmd(),
		newJwtVerifyCmd(),
	)

	return cmd
}

func newJwtInspectCmd() *cobra.Command {
	var bxtOut string

	cmd := &cobra.Command{
		Use:     "inspect [token]",
		Aliases: []string{"decode"},
		Short:   "Decode and inspect a JWT header and payload without verifying",
		RunE: func(cmd *cobra.Command, args []string) error {
			in, err := readInput(cmd, args)
			if err != nil {
				return err
			}
			tokenStr := strings.TrimSpace(string(in))
			parts := strings.Split(tokenStr, ".")
			if len(parts) != 3 {
				return fmt.Errorf("invalid JWT: expected 3 parts separated by dots, got %d", len(parts))
			}

			headerRaw, err := cryptzb64.UrlDecode(parts[0])
			if err != nil {
				return fmt.Errorf("invalid header base64: %w", err)
			}
			payloadRaw, err := cryptzb64.UrlDecode(parts[1])
			if err != nil {
				return fmt.Errorf("invalid payload base64: %w", err)
			}

			var headerMap any
			if err := json.Unmarshal(headerRaw.Bytes(), &headerMap); err != nil {
				return fmt.Errorf("invalid header JSON: %w", err)
			}

			var payloadMap any
			if err := json.Unmarshal(payloadRaw.Bytes(), &payloadMap); err != nil {
				return fmt.Errorf("invalid payload JSON: %w", err)
			}

			var payloadObj cryptzjwt.JWTPayload
			_ = json.Unmarshal(payloadRaw.Bytes(), &payloadObj)

			if bxtOut != "" {
				if len(payloadObj.Bxt) == 0 {
					return fmt.Errorf("token does not contain a binary extension (bxt)")
				}
				if err := os.WriteFile(bxtOut, payloadObj.Bxt, 0644); err != nil {
					return fmt.Errorf("failed to write bxt to %s: %w", bxtOut, err)
				}
			}

			result := struct {
				Header    any    `json:"header"`
				Payload   any    `json:"payload"`
				Signature string `json:"signature"`
			}{
				Header:    headerMap,
				Payload:   payloadMap,
				Signature: parts[2],
			}

			formatted, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(formatted))
			return nil
		},
	}

	cmd.Flags().StringVar(&bxtOut, "bxt-out", "", "File path to write raw binary extension (bxt) contents")

	return cmd
}

func newJwtSealCmd() *cobra.Command {
	var keyPath string
	var alg string
	var iss string
	var sub string
	var aud string
	var expStr string
	var typ string
	var ext string
	var idp string
	var jti string
	var iatStr string
	var bxt string
	var bxtHex string
	var bxtB64 string
	var bxtB64Url string
	var bxtFile string

	cmd := &cobra.Command{
		Use:     "seal",
		Aliases: []string{"create"},
		Short:   "Create and sign an RS256 or EdDSA JWT token using a private key",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyPEM, err := loadKey(keyPath)
			if err != nil {
				return err
			}

			expDuration, err := time.ParseDuration(expStr)
			if err != nil {
				return fmt.Errorf("invalid duration format for --exp: %w", err)
			}

			iatTime, err := parseIat(iatStr)
			if err != nil {
				return err
			}

			bxtBytes, err := parseBxt(bxt, bxtHex, bxtB64, bxtB64Url, bxtFile)
			if err != nil {
				return err
			}

			payload, err := buildJwtPayload(jti, typ, iatTime, expDuration, iss, aud, sub, ext, bxtBytes, idp)
			if err != nil {
				return fmt.Errorf("failed to build payload: %w", err)
			}

			var tokenStr string
			switch strings.ToUpper(alg) {
			case "EDDSA":
				s := cryptz25519.NewService()
				priv, err := safePrivKeyImport25519(s, keyPEM)
				if err != nil {
					return fmt.Errorf("failed to import Ed25519 private key: %w", err)
				}
				kr := cryptz25519keyring.NewKeyRing()
				kr.AddPrivate(iss, priv)
				_, tokenStr = cryptzjwt.Seal25519(kr, payload)
			case "RS256":
				s := cryptzrsa.NewService()
				priv, err := safePrivKeyImport(s, keyPEM)
				if err != nil {
					return fmt.Errorf("failed to import RSA private key: %w", err)
				}
				kr := cryptzrsakeyring.NewKeyRing()
				kr.AddPrivate(iss, priv)
				_, tokenStr = cryptzjwt.Seal(kr, payload)
			default:
				return fmt.Errorf("unsupported algorithm %s: expected RS256 or EdDSA", alg)
			}

			fmt.Fprintln(cmd.OutOrStdout(), tokenStr)
			return nil
		},
	}

	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "Private key file or PEM string (use '-' for stdin)")
	cmd.Flags().StringVar(&alg, "alg", "RS256", "Signature algorithm: RS256 or EdDSA")
	cmd.Flags().StringVar(&iss, "iss", "app", "Issuer (must match ^[a-z][a-z0-9]*$)")
	cmd.Flags().StringVar(&sub, "sub", "", "Subject")
	cmd.Flags().StringVar(&aud, "aud", "", "Audience")
	cmd.Flags().StringVar(&expStr, "exp", "1h", "Expiration duration (e.g. 15m, 1h, 24h)")
	cmd.Flags().StringVar(&typ, "typ", "", "Payload type")
	cmd.Flags().StringVar(&ext, "ext", "", "String extension")
	cmd.Flags().StringVar(&idp, "idp", "", "Identity provider")
	cmd.Flags().StringVar(&jti, "jti", "", "JWT ID (optional, generated if not set)")
	cmd.Flags().StringVar(&iatStr, "iat", "", "Issued at timestamp (unix seconds/ms, RFC3339, or relative offset like -5m)")
	cmd.Flags().StringVar(&bxt, "bxt", "", "Binary extension as raw string")
	cmd.Flags().StringVar(&bxtHex, "bxt-hex", "", "Binary extension as hex string")
	cmd.Flags().StringVar(&bxtB64, "bxt-b64", "", "Binary extension as Base64 string")
	cmd.Flags().StringVar(&bxtB64Url, "bxt-b64url", "", "Binary extension as URL-safe Base64 string")
	cmd.Flags().StringVar(&bxtFile, "bxt-file", "", "Binary extension read from file")

	cmd.MarkFlagsMutuallyExclusive("bxt", "bxt-hex", "bxt-b64", "bxt-b64url", "bxt-file")

	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("sub")
	cmd.MarkFlagRequired("aud")

	return cmd
}

func newJwtVerifyCmd() *cobra.Command {
	var keyPath string
	var expectedIss string
	var expectedAud string
	var expectedSub string
	var bxtOut string

	cmd := &cobra.Command{
		Use:     "verify [token]",
		Aliases: []string{"unseal"},
		Short:   "Verify an RS256 or EdDSA JWT signature and claims using a public key",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyPEM, err := loadKey(keyPath)
			if err != nil {
				return err
			}

			in, err := readInput(cmd, args)
			if err != nil {
				return err
			}
			tokenStr := strings.TrimSpace(string(in))

			parts := strings.Split(tokenStr, ".")
			if len(parts) != 3 {
				return fmt.Errorf("invalid token format: expected 3 parts, got %d", len(parts))
			}

			headerRaw, err := cryptzb64.UrlDecode(parts[0])
			if err != nil {
				return fmt.Errorf("invalid header base64: %w", err)
			}
			var header cryptzjwt.JWTHeader
			if err := json.Unmarshal(headerRaw.Bytes(), &header); err != nil {
				return fmt.Errorf("invalid header JSON: %w", err)
			}

			var token *cryptzjwt.JWTToken
			if strings.EqualFold(header.Alg, "EDDSA") {
				s := cryptz25519.NewService()
				pub, err := safePubKeyImport25519(s, keyPEM)
				if err != nil {
					priv, privErr := safePrivKeyImport25519(s, keyPEM)
					if privErr != nil {
						return fmt.Errorf("failed to parse Ed25519 key: %w", err)
					}
					pub = cryptz25519.ExtractPubKey(priv)
				}
				token, err = safeUnseal25519(tokenStr, pub)
				if err != nil {
					fmt.Fprintln(cmd.OutOrStderr(), "Verification failed:", err)
					return fmt.Errorf("verification failed: %w", err)
				}
			} else {
				s := cryptzrsa.NewService()
				pub, err := safePubKeyImport(s, keyPEM)
				if err != nil {
					priv, privErr := safePrivKeyImport(s, keyPEM)
					if privErr != nil {
						return fmt.Errorf("failed to parse RSA key: %w", err)
					}
					pub = &priv.PublicKey
				}
				token, err = safeUnseal(tokenStr, pub)
				if err != nil {
					fmt.Fprintln(cmd.OutOrStderr(), "Verification failed:", err)
					return fmt.Errorf("verification failed: %w", err)
				}
			}

			if expectedIss != "" && token.Payload.Iss != expectedIss {
				return fmt.Errorf("issuer mismatch: expected %s, got %s", expectedIss, token.Payload.Iss)
			}
			if expectedAud != "" && token.Payload.Aud != expectedAud {
				return fmt.Errorf("audience mismatch: expected %s, got %s", expectedAud, token.Payload.Aud)
			}
			if expectedSub != "" && token.Payload.Sub != expectedSub {
				return fmt.Errorf("subject mismatch: expected %s, got %s", expectedSub, token.Payload.Sub)
			}

			if bxtOut != "" {
				if len(token.Payload.Bxt) == 0 {
					return fmt.Errorf("token does not contain a binary extension (bxt)")
				}
				if err := os.WriteFile(bxtOut, token.Payload.Bxt, 0644); err != nil {
					return fmt.Errorf("failed to write bxt to %s: %w", bxtOut, err)
				}
			}

			formatted, err := json.MarshalIndent(token, "", "  ")
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(formatted))
			return nil
		},
	}

	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "Public (or private) key file or PEM string (use '-' for stdin)")
	cmd.Flags().StringVar(&expectedIss, "iss", "", "Assert expected issuer")
	cmd.Flags().StringVar(&expectedAud, "aud", "", "Assert expected audience")
	cmd.Flags().StringVar(&expectedSub, "sub", "", "Assert expected subject")
	cmd.Flags().StringVar(&bxtOut, "bxt-out", "", "File path to write raw binary extension (bxt) contents upon successful verification")
	cmd.MarkFlagRequired("key")

	return cmd
}

func safeUnseal(token string, key *rsa.PublicKey) (tk *cryptzjwt.JWTToken, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return cryptzjwt.UnsealCallback(token, func(id, iss string) (*rsa.PublicKey, error) {
		return key, nil
	})
}

func safeUnseal25519(token string, key ed25519.PublicKey) (tk *cryptzjwt.JWTToken, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return cryptzjwt.Unseal25519Callback(token, func(id, iss string) (ed25519.PublicKey, error) {
		return key, nil
	})
}
