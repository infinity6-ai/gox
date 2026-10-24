package cmdcryptz

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/infinity6-ai/gox/cryptz/cryptzaes"
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

func newJwtStateEncryptCmd() *cobra.Command {
	var keyHex string
	var krPath string
	var src string
	var dst string
	var keyId string
	var b64 bool
	var b64url bool
	var raw bool

	cmd := &cobra.Command{
		Use:     "encrypt [message]",
		Aliases: []string{"send"},
		Short:   "Encrypt state data into a symmetric envelope using AES-GCM",
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

			if keyHex != "" {
				keyBytes, err := hex.DecodeString(cleanHexInput([]byte(keyHex)))
				if err != nil {
					return fmt.Errorf("invalid key hex: %w", err)
				}
				keysLoader = &cryptzjwtstate.EncryptorKeysLoader{
					KeyLoader: func(s, d string) cryptzaes.Key {
						return cryptzaes.Key(keyBytes)
					},
				}
			} else if krPath != "" {
				kr, err := loadStateKeyRingFile(krPath)
				if err != nil {
					return err
				}
				keysLoader = cryptzjwtstate.FromKeyRing(kr).EncryptorKeysLoader()
			} else {
				return fmt.Errorf("either --key or --kr must be provided")
			}

			enc := cryptzjwtstate.NewEncryptor()
			enc.KeysLoader = keysLoader
			enc.Message = &cryptzjwtstate.StateMessagePlain{
				SrcKey: src,
				DstKey: dst,
				KeyId:  keyId,
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
	cmd.Flags().StringVar(&krPath, "kr", "", "State keyring file path")
	cmd.Flags().StringVar(&src, "src", "app", "Source / owner name")
	cmd.Flags().StringVar(&dst, "dst", "", "Destination name (defaults to src)")
	cmd.Flags().StringVar(&keyId, "key-id", "", "Optional key identifier")
	cmd.Flags().BoolVar(&b64, "b64", false, "Output in Base64")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Output in URL-safe Base64")
	cmd.Flags().BoolVarP(&raw, "raw", "r", false, "Output raw binary")

	cmd.MarkFlagsMutuallyExclusive("key", "kr")
	cmd.MarkFlagsMutuallyExclusive("b64", "b64url", "raw")

	return cmd
}

func newJwtStateDecryptCmd() *cobra.Command {
	var keyHex string
	var krPath string
	var b64 bool
	var b64url bool
	var raw bool

	cmd := &cobra.Command{
		Use:     "decrypt [ciphertext]",
		Aliases: []string{"receive"},
		Short:   "Decrypt a symmetric state envelope using AES-GCM",
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

			if keyHex != "" {
				keyBytes, err := hex.DecodeString(cleanHexInput([]byte(keyHex)))
				if err != nil {
					return fmt.Errorf("invalid key hex: %w", err)
				}
				keysLoader = &cryptzjwtstate.DecryptorKeysLoader{
					KeyLoader: func(s, d string) cryptzaes.Key {
						return cryptzaes.Key(keyBytes)
					},
				}
			} else if krPath != "" {
				kr, err := loadStateKeyRingFile(krPath)
				if err != nil {
					return err
				}
				keysLoader = cryptzjwtstate.FromKeyRing(kr).DecryptorKeysLoader()
			} else {
				return fmt.Errorf("either --key or --kr must be provided")
			}

			dec := cryptzjwtstate.NewDecryptor()
			dec.KeysLoader = keysLoader
			dec.CipheredData = cipherBytes

			plain, err := safeStateDecrypt(dec)
			if err != nil {
				fmt.Fprintln(cmd.OutOrStderr(), "State decryption failed:", err)
				return fmt.Errorf("state decryption failed: %w", err)
			}

			_, err = cmd.OutOrStdout().Write(plain.Data)
			return err
		},
	}

	cmd.Flags().StringVarP(&keyHex, "key", "k", "", "Direct AES symmetric key in hex")
	cmd.Flags().StringVar(&krPath, "kr", "", "State keyring file path")
	cmd.Flags().BoolVar(&b64, "b64", false, "Input ciphertext is Base64")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Input ciphertext is URL-safe Base64")
	cmd.Flags().BoolVarP(&raw, "raw", "r", false, "Input ciphertext is raw binary")

	cmd.MarkFlagsMutuallyExclusive("key", "kr")
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
