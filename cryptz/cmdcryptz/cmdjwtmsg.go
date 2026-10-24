package cmdcryptz

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt/cryptzjwtmsg"
	"github.com/spf13/cobra"
)

func newJwtMsgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jwtmsg",
		Aliases: []string{"msg"},
		Short:   "Hybrid RSA/AES messaging between parties using KeyRing",
	}

	cmd.AddCommand(
		newJwtMsgEncryptCmd(),
		newJwtMsgDecryptCmd(),
		newJwtMsgSendCmd(),
		newJwtMsgReceiveCmd(),
	)

	return cmd
}

func newJwtMsgEncryptCmd() *cobra.Command {
	var krPath string
	var src string
	var dst string
	var b64 bool
	var b64url bool
	var raw bool

	cmd := &cobra.Command{
		Use:   "encrypt [message]",
		Short: "Encrypt and sign raw data from sender to recipient using KeyRing",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadKeyRingFile(krPath)
			if err != nil {
				return err
			}

			in, err := readInput(cmd, args)
			if err != nil {
				return err
			}

			mkr := cryptzjwtmsg.FromKeyRing(kr)
			enc := cryptzjwtmsg.NewEncryptor()
			enc.KeysLoader = mkr.EncryptorKeysLoader()
			enc.Message = &cryptzjwtmsg.RSAMessagePlain{
				SrcKey: src,
				DstKey: dst,
				Data:   in,
			}

			ciphertext, err := safeMsgEncrypt(enc)
			if err != nil {
				return err
			}

			return outputBytes(cmd, ciphertext, b64, b64url, raw)
		},
	}

	cmd.Flags().StringVarP(&krPath, "kr", "k", "keyring.json", "Keyring file path")
	cmd.Flags().StringVar(&src, "src", "", "Sender owner name in keyring (must have private key to sign)")
	cmd.Flags().StringVar(&dst, "dst", "", "Recipient owner name in keyring (must have public key to encrypt)")
	cmd.Flags().BoolVar(&b64, "b64", false, "Output in Base64")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Output in URL-safe Base64")
	cmd.Flags().BoolVarP(&raw, "raw", "r", false, "Output raw binary")

	cmd.MarkFlagsMutuallyExclusive("b64", "b64url", "raw")
	cmd.MarkFlagRequired("src")
	cmd.MarkFlagRequired("dst")

	return cmd
}

func newJwtMsgDecryptCmd() *cobra.Command {
	var krPath string
	var expectedSrc string
	var expectedDst string
	var b64 bool
	var b64url bool
	var raw bool

	cmd := &cobra.Command{
		Use:   "decrypt [ciphertext]",
		Short: "Decrypt and verify raw data using recipient private key and sender public key from KeyRing",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadKeyRingFile(krPath)
			if err != nil {
				return err
			}

			in, err := readInput(cmd, args)
			if err != nil {
				return err
			}

			var cipherBytes []byte
			if raw {
				cipherBytes = in
			} else {
				var err error
				cipherBytes, err = decodeInputBytes(in, b64, b64url)
				if err != nil {
					return err
				}
			}

			mkr := cryptzjwtmsg.FromKeyRing(kr)
			dec := cryptzjwtmsg.NewDecryptor()
			dec.KeysLoader = mkr.DecryptorKeysLoader()
			dec.CipheredData = cipherBytes

			plain, err := safeMsgDecrypt(dec)
			if err != nil {
				fmt.Fprintln(cmd.OutOrStderr(), "Decryption failed:", err)
				return fmt.Errorf("decryption failed: %w", err)
			}

			if expectedSrc != "" && plain.SrcKey != expectedSrc {
				return fmt.Errorf("source mismatch: expected %s, got %s", expectedSrc, plain.SrcKey)
			}
			if expectedDst != "" && plain.DstKey != expectedDst {
				return fmt.Errorf("destination mismatch: expected %s, got %s", expectedDst, plain.DstKey)
			}

			_, err = cmd.OutOrStdout().Write(plain.Data)
			return err
		},
	}

	cmd.Flags().StringVarP(&krPath, "kr", "k", "keyring.json", "Keyring file path")
	cmd.Flags().StringVar(&expectedSrc, "src", "", "Assert expected sender/source owner name")
	cmd.Flags().StringVar(&expectedDst, "dst", "", "Assert expected recipient/destination owner name")
	cmd.Flags().BoolVar(&b64, "b64", false, "Input ciphertext is Base64")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Input ciphertext is URL-safe Base64")
	cmd.Flags().BoolVarP(&raw, "raw", "r", false, "Input ciphertext is raw binary")
	cmd.MarkFlagsMutuallyExclusive("b64", "b64url", "raw")

	return cmd
}

func newJwtMsgSendCmd() *cobra.Command {
	var krPath string
	var src string
	var dst string
	var sub string
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
	var b64 bool
	var b64url bool
	var raw bool

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a structured JWT payload envelope to recipient (enforces exp, iat, claims)",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadKeyRingFile(krPath)
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

			payload, err := buildJwtPayload(jti, typ, iatTime, expDuration, src, dst, sub, ext, bxtBytes, idp)
			if err != nil {
				return fmt.Errorf("failed to build payload: %w", err)
			}

			mkr := cryptzjwtmsg.FromKeyRing(kr)
			ciphertext, err := safeJwtMsgSend(payload, mkr.EncryptorKeysLoader())
			if err != nil {
				return err
			}

			return outputBytes(cmd, ciphertext, b64, b64url, raw)
		},
	}

	cmd.Flags().StringVarP(&krPath, "kr", "k", "keyring.json", "Keyring file path")
	cmd.Flags().StringVar(&src, "src", "", "Sender owner name (sets iss)")
	cmd.Flags().StringVar(&dst, "dst", "", "Recipient owner name (sets aud)")
	cmd.Flags().StringVar(&sub, "sub", "", "Subject claim")
	cmd.Flags().StringVar(&expStr, "exp", "1h", "Expiration duration (e.g. 15m, 1h, 24h)")
	cmd.Flags().StringVar(&iatStr, "iat", "", "Issued at timestamp (unix seconds/ms, RFC3339, or relative offset like -5m)")
	cmd.Flags().StringVar(&typ, "typ", "", "Payload type")
	cmd.Flags().StringVar(&ext, "ext", "", "String extension")
	cmd.Flags().StringVar(&idp, "idp", "", "Identity provider")
	cmd.Flags().StringVar(&jti, "jti", "", "JWT ID (optional, generated if not set)")
	cmd.Flags().StringVar(&bxt, "bxt", "", "Binary extension as raw string")
	cmd.Flags().StringVar(&bxtHex, "bxt-hex", "", "Binary extension as hex string")
	cmd.Flags().StringVar(&bxtB64, "bxt-b64", "", "Binary extension as Base64 string")
	cmd.Flags().StringVar(&bxtB64Url, "bxt-b64url", "", "Binary extension as URL-safe Base64 string")
	cmd.Flags().StringVar(&bxtFile, "bxt-file", "", "Binary extension read from file")
	cmd.Flags().BoolVar(&b64, "b64", false, "Output in Base64")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Output in URL-safe Base64")
	cmd.Flags().BoolVarP(&raw, "raw", "r", false, "Output raw binary")

	cmd.MarkFlagsMutuallyExclusive("bxt", "bxt-hex", "bxt-b64", "bxt-b64url", "bxt-file")
	cmd.MarkFlagsMutuallyExclusive("b64", "b64url", "raw")
	cmd.MarkFlagRequired("src")
	cmd.MarkFlagRequired("dst")
	cmd.MarkFlagRequired("sub")

	return cmd
}

func newJwtMsgReceiveCmd() *cobra.Command {
	var krPath string
	var expectedSrc string
	var expectedDst string
	var expectedSub string
	var bxtOut string
	var b64 bool
	var b64url bool
	var raw bool

	cmd := &cobra.Command{
		Use:   "receive [ciphertext]",
		Short: "Receive and verify a structured JWT payload envelope (verifies exp, iat, claims)",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadKeyRingFile(krPath)
			if err != nil {
				return err
			}

			in, err := readInput(cmd, args)
			if err != nil {
				return err
			}

			var cipherBytes []byte
			if raw {
				cipherBytes = in
			} else {
				var err error
				cipherBytes, err = decodeInputBytes(in, b64, b64url)
				if err != nil {
					return err
				}
			}

			mkr := cryptzjwtmsg.FromKeyRing(kr)
			payload, err := safeJwtMsgReceive(cipherBytes, mkr.DecryptorKeysLoader())
			if err != nil {
				fmt.Fprintln(cmd.OutOrStderr(), "Receive verification failed:", err)
				return fmt.Errorf("verification failed: %w", err)
			}

			if expectedSrc != "" && payload.Iss != expectedSrc {
				return fmt.Errorf("source mismatch: expected %s, got %s", expectedSrc, payload.Iss)
			}
			if expectedDst != "" && payload.Aud != expectedDst {
				return fmt.Errorf("destination mismatch: expected %s, got %s", expectedDst, payload.Aud)
			}
			if expectedSub != "" && payload.Sub != expectedSub {
				return fmt.Errorf("subject mismatch: expected %s, got %s", expectedSub, payload.Sub)
			}

			if bxtOut != "" {
				if len(payload.Bxt) == 0 {
					return fmt.Errorf("payload does not contain a binary extension (bxt)")
				}
				if err := os.WriteFile(bxtOut, payload.Bxt, 0644); err != nil {
					return fmt.Errorf("failed to write bxt to %s: %w", bxtOut, err)
				}
			}

			formatted, err := json.MarshalIndent(payload, "", "  ")
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(formatted))
			return nil
		},
	}

	cmd.Flags().StringVarP(&krPath, "kr", "k", "keyring.json", "Keyring file path")
	cmd.Flags().StringVar(&expectedSrc, "src", "", "Assert expected sender/source owner name")
	cmd.Flags().StringVar(&expectedDst, "dst", "", "Assert expected recipient/destination owner name")
	cmd.Flags().StringVar(&expectedSub, "sub", "", "Assert expected subject claim")
	cmd.Flags().StringVar(&bxtOut, "bxt-out", "", "File path to write raw binary extension (bxt) contents upon successful verification")
	cmd.Flags().BoolVar(&b64, "b64", false, "Input ciphertext is Base64")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Input ciphertext is URL-safe Base64")
	cmd.Flags().BoolVarP(&raw, "raw", "r", false, "Input ciphertext is raw binary")
	cmd.MarkFlagsMutuallyExclusive("b64", "b64url", "raw")

	return cmd
}

func safeMsgEncrypt(enc *cryptzjwtmsg.Encryptor) (out []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("encrypt failed: %v", r)
		}
	}()
	return enc.Encrypt(), nil
}

func safeMsgDecrypt(dec *cryptzjwtmsg.Decryptor) (out *cryptzjwtmsg.RSAMessagePlain, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("decrypt failed: %v", r)
		}
	}()
	return dec.Decrypt(), nil
}

func safeJwtMsgSend(payload *cryptzjwt.JWTPayload, keysLoader *cryptzjwtmsg.EncryptorKeysLoader) (out []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("send failed: %v", r)
		}
	}()
	return cryptzjwtmsg.Send(payload, keysLoader), nil
}

func safeJwtMsgReceive(ciphertext []byte, keysLoader *cryptzjwtmsg.DecryptorKeysLoader) (payload *cryptzjwt.JWTPayload, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return cryptzjwtmsg.Receive(ciphertext, keysLoader), nil
}
