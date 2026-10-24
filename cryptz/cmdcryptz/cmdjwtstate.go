package cmdcryptz

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/infinity6-ai/gox/cryptz/cryptzaes"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt/cryptzjwtstate"
	"github.com/spf13/cobra"
)

func newJwtStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jwtstate",
		Aliases: []string{"state"},
		Short:   "Symmetric AES state envelopes (ideal for session/cookie/state tokens)",
	}

	cmd.AddCommand(
		newJwtStateEncryptCmd(),
		newJwtStateDecryptCmd(),
		newJwtStateSendCmd(),
		newJwtStateReceiveCmd(),
	)

	return cmd
}

func loadStateKeyRingFile(filePath string) (*cryptzjwtstate.KeyRing, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state keyring file %s: %w", filePath, err)
	}
	kr, err := cryptzjwtstate.Import(data)
	if err != nil {
		return nil, fmt.Errorf("failed to import state keyring from %s: %w", filePath, err)
	}
	return kr, nil
}

func loadAesKey(keyHex string, keyFile string) (cryptzaes.Key, error) {
	if keyHex != "" {
		keyBytes, err := hex.DecodeString(cleanHexInput([]byte(keyHex)))
		if err != nil {
			return nil, fmt.Errorf("invalid key hex: %w", err)
		}
		return cryptzaes.Key(keyBytes), nil
	}
	if keyFile != "" {
		data, err := os.ReadFile(keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read key file %s: %w", keyFile, err)
		}
		str := cleanHexInput(data)
		if keyBytes, err := hex.DecodeString(str); err == nil && (len(keyBytes) == 16 || len(keyBytes) == 24 || len(keyBytes) == 32) {
			return cryptzaes.Key(keyBytes), nil
		}
		if len(data) == 16 || len(data) == 24 || len(data) == 32 {
			return cryptzaes.Key(data), nil
		}
		return nil, fmt.Errorf("key file %s must contain a 16, 24, or 32 byte key (in hex or raw binary)", keyFile)
	}
	return nil, nil
}

func newJwtStateEncryptCmd() *cobra.Command {
	var keyHex string
	var keyFile string
	var krPath string
	var src string
	var dst string
	var b64 bool
	var b64url bool
	var raw bool

	cmd := &cobra.Command{
		Use:   "encrypt [message]",
		Short: "Encrypt raw state data into a symmetric envelope using AES-GCM",
		RunE: func(cmd *cobra.Command, args []string) error {
			if src == "" {
				src = "app"
			}
			if dst == "" {
				dst = src
			}

			in, err := readInput(cmd, args)
			if err != nil {
				return err
			}

			var keysLoader *cryptzjwtstate.EncryptorKeysLoader

			directKey, err := loadAesKey(keyHex, keyFile)
			if err != nil {
				return err
			}

			if directKey != nil {
				keysLoader = &cryptzjwtstate.EncryptorKeysLoader{
					KeyLoader: func(s, d string) cryptzaes.Key {
						return directKey
					},
				}
			} else if krPath != "" {
				kr, err := loadStateKeyRingFile(krPath)
				if err != nil {
					return err
				}
				keysLoader = cryptzjwtstate.FromKeyRing(kr).EncryptorKeysLoader()
			} else {
				return fmt.Errorf("one of --key, --key-file, or --kr must be provided")
			}

			enc := cryptzjwtstate.NewEncryptor()
			enc.KeysLoader = keysLoader
			enc.Message = &cryptzjwtstate.StateMessagePlain{
				SrcKey: src,
				DstKey: dst,
				Data:   in,
			}

			ciphertext, err := safeStateEncrypt(enc)
			if err != nil {
				return err
			}

			return outputBytes(cmd, ciphertext, b64, b64url, raw)
		},
	}

	cmd.Flags().StringVarP(&keyHex, "key", "k", "", "Direct AES symmetric key in hex (16, 24, or 32 bytes)")
	cmd.Flags().StringVar(&keyFile, "key-file", "", "File containing AES symmetric key (hex or raw binary)")
	cmd.Flags().StringVar(&krPath, "kr", "", "State keyring file path")
	cmd.Flags().StringVar(&src, "src", "app", "Source / owner name")
	cmd.Flags().StringVar(&dst, "dst", "", "Destination name (defaults to src)")
	cmd.Flags().BoolVar(&b64, "b64", false, "Output in Base64")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Output in URL-safe Base64")
	cmd.Flags().BoolVarP(&raw, "raw", "r", false, "Output raw binary")

	cmd.MarkFlagsMutuallyExclusive("key", "key-file", "kr")
	cmd.MarkFlagsMutuallyExclusive("b64", "b64url", "raw")

	return cmd
}

func newJwtStateDecryptCmd() *cobra.Command {
	var keyHex string
	var keyFile string
	var krPath string
	var expectedSrc string
	var expectedDst string
	var b64 bool
	var b64url bool
	var raw bool

	cmd := &cobra.Command{
		Use:   "decrypt [ciphertext]",
		Short: "Decrypt a raw symmetric state envelope using AES-GCM",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			var keysLoader *cryptzjwtstate.DecryptorKeysLoader

			directKey, err := loadAesKey(keyHex, keyFile)
			if err != nil {
				return err
			}

			if directKey != nil {
				keysLoader = &cryptzjwtstate.DecryptorKeysLoader{
					KeyLoader: func(s, d string) cryptzaes.Key {
						return directKey
					},
				}
			} else if krPath != "" {
				kr, err := loadStateKeyRingFile(krPath)
				if err != nil {
					return err
				}
				keysLoader = cryptzjwtstate.FromKeyRing(kr).DecryptorKeysLoader()
			} else {
				return fmt.Errorf("one of --key, --key-file, or --kr must be provided")
			}

			dec := cryptzjwtstate.NewDecryptor()
			dec.KeysLoader = keysLoader
			dec.CipheredData = cipherBytes

			plain, err := safeStateDecrypt(dec)
			if err != nil {
				fmt.Fprintln(cmd.OutOrStderr(), "State decryption failed:", err)
				return fmt.Errorf("state decryption failed: %w", err)
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

	cmd.Flags().StringVarP(&keyHex, "key", "k", "", "Direct AES symmetric key in hex")
	cmd.Flags().StringVar(&keyFile, "key-file", "", "File containing AES symmetric key (hex or raw binary)")
	cmd.Flags().StringVar(&krPath, "kr", "", "State keyring file path")
	cmd.Flags().StringVar(&expectedSrc, "src", "", "Assert expected source owner name")
	cmd.Flags().StringVar(&expectedDst, "dst", "", "Assert expected destination name")
	cmd.Flags().BoolVar(&b64, "b64", false, "Input ciphertext is Base64")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Input ciphertext is URL-safe Base64")
	cmd.Flags().BoolVarP(&raw, "raw", "r", false, "Input ciphertext is raw binary")

	cmd.MarkFlagsMutuallyExclusive("key", "key-file", "kr")
	cmd.MarkFlagsMutuallyExclusive("b64", "b64url", "raw")

	return cmd
}

func newJwtStateSendCmd() *cobra.Command {
	var keyHex string
	var keyFile string
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
		Short: "Send a structured JWT payload state envelope (enforces exp, iat, claims)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if src == "" {
				src = "app"
			}
			if dst == "" {
				dst = src
			}

			directKey, err := loadAesKey(keyHex, keyFile)
			if err != nil {
				return err
			}

			var keysLoader *cryptzjwtstate.EncryptorKeysLoader
			if directKey != nil {
				keysLoader = &cryptzjwtstate.EncryptorKeysLoader{
					KeyLoader: func(s, d string) cryptzaes.Key {
						return directKey
					},
				}
			} else if krPath != "" {
				kr, err := loadStateKeyRingFile(krPath)
				if err != nil {
					return err
				}
				keysLoader = cryptzjwtstate.FromKeyRing(kr).EncryptorKeysLoader()
			} else {
				return fmt.Errorf("one of --key, --key-file, or --kr must be provided")
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

			ciphertext, err := safeJwtStateSend(payload, keysLoader)
			if err != nil {
				return err
			}

			return outputBytes(cmd, ciphertext, b64, b64url, raw)
		},
	}

	cmd.Flags().StringVarP(&keyHex, "key", "k", "", "Direct AES symmetric key in hex")
	cmd.Flags().StringVar(&keyFile, "key-file", "", "File containing AES symmetric key (hex or raw binary)")
	cmd.Flags().StringVar(&krPath, "kr", "", "State keyring file path")
	cmd.Flags().StringVar(&src, "src", "app", "Source / owner name (sets iss)")
	cmd.Flags().StringVar(&dst, "dst", "", "Destination name (defaults to src, sets aud)")
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

	cmd.MarkFlagsMutuallyExclusive("key", "key-file", "kr")
	cmd.MarkFlagsMutuallyExclusive("bxt", "bxt-hex", "bxt-b64", "bxt-b64url", "bxt-file")
	cmd.MarkFlagsMutuallyExclusive("b64", "b64url", "raw")
	cmd.MarkFlagRequired("sub")

	return cmd
}

func newJwtStateReceiveCmd() *cobra.Command {
	var keyHex string
	var keyFile string
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
		Short: "Receive and verify a structured JWT payload state envelope (verifies exp, iat, claims)",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			directKey, err := loadAesKey(keyHex, keyFile)
			if err != nil {
				return err
			}

			var keysLoader *cryptzjwtstate.DecryptorKeysLoader
			if directKey != nil {
				keysLoader = &cryptzjwtstate.DecryptorKeysLoader{
					KeyLoader: func(s, d string) cryptzaes.Key {
						return directKey
					},
				}
			} else if krPath != "" {
				kr, err := loadStateKeyRingFile(krPath)
				if err != nil {
					return err
				}
				keysLoader = cryptzjwtstate.FromKeyRing(kr).DecryptorKeysLoader()
			} else {
				return fmt.Errorf("one of --key, --key-file, or --kr must be provided")
			}

			payload, err := safeJwtStateReceive(cipherBytes, keysLoader)
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

	cmd.Flags().StringVarP(&keyHex, "key", "k", "", "Direct AES symmetric key in hex")
	cmd.Flags().StringVar(&keyFile, "key-file", "", "File containing AES symmetric key (hex or raw binary)")
	cmd.Flags().StringVar(&krPath, "kr", "", "State keyring file path")
	cmd.Flags().StringVar(&expectedSrc, "src", "", "Assert expected source owner name")
	cmd.Flags().StringVar(&expectedDst, "dst", "", "Assert expected destination name")
	cmd.Flags().StringVar(&expectedSub, "sub", "", "Assert expected subject claim")
	cmd.Flags().StringVar(&bxtOut, "bxt-out", "", "File path to write raw binary extension (bxt) contents upon successful verification")
	cmd.Flags().BoolVar(&b64, "b64", false, "Input ciphertext is Base64")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Input ciphertext is URL-safe Base64")
	cmd.Flags().BoolVarP(&raw, "raw", "r", false, "Input ciphertext is raw binary")

	cmd.MarkFlagsMutuallyExclusive("key", "key-file", "kr")
	cmd.MarkFlagsMutuallyExclusive("b64", "b64url", "raw")

	return cmd
}

func safeStateEncrypt(enc *cryptzjwtstate.Encryptor) (out []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("encrypt failed: %v", r)
		}
	}()
	return enc.Encrypt(), nil
}

func safeStateDecrypt(dec *cryptzjwtstate.Decryptor) (out *cryptzjwtstate.StateMessagePlain, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("decrypt failed: %v", r)
		}
	}()
	return dec.Decrypt(), nil
}

func safeJwtStateSend(payload *cryptzjwt.JWTPayload, keysLoader *cryptzjwtstate.EncryptorKeysLoader) (out []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("send failed: %v", r)
		}
	}()
	return cryptzjwtstate.Send(payload, keysLoader), nil
}

func safeJwtStateReceive(ciphertext []byte, keysLoader *cryptzjwtstate.DecryptorKeysLoader) (payload *cryptzjwt.JWTPayload, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return cryptzjwtstate.Receive(ciphertext, keysLoader), nil
}
