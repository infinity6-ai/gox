package cmdcryptz

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/infinity6-ai/gox/cryptz/cryptz25519"
	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519service"
	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
	"github.com/spf13/cobra"
)

func new25519Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "25519",
		Short: "Curve25519 (Ed25519/X25519) utilities (keygen, pubkey extraction, sign, verify, encrypt, decrypt, id)",
	}

	cmd.AddCommand(
		new25519GenCmd(),
		new25519PubCmd(),
		new25519SignCmd(),
		new25519VerifyCmd(),
		new25519EncryptCmd(),
		new25519DecryptCmd(),
		new25519IdCmd(),
	)

	return cmd
}

func new25519GenCmd() *cobra.Command {
	var privOut string
	var pubOut string

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate Curve25519 key pair",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := cryptz25519.NewService()
			priv := s.PrivKeyCreate()
			pub := priv.Public().(ed25519.PublicKey)
			privStr := s.PrivKeyExport(priv)
			pubStr := s.PubKeyExport(pub)
			stdout := cmd.OutOrStdout()

			if privOut != "" {
				if err := os.WriteFile(privOut, []byte(privStr), 0600); err != nil {
					return fmt.Errorf("failed to write private key to %s: %w", privOut, err)
				}
				fmt.Fprintf(stdout, "Private key written to %s\n", privOut)
			} else {
				fmt.Fprintln(stdout, privStr)
			}

			if pubOut != "" {
				if err := os.WriteFile(pubOut, []byte(pubStr), 0644); err != nil {
					return fmt.Errorf("failed to write public key to %s: %w", pubOut, err)
				}
				fmt.Fprintf(stdout, "Public key written to %s\n", pubOut)
			} else {
				fmt.Fprintln(stdout, pubStr)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&privOut, "priv-out", "", "File to write private key to")
	cmd.Flags().StringVar(&pubOut, "pub-out", "", "File to write public key to")

	return cmd
}

func new25519PubCmd() *cobra.Command {
	var keyPath string
	var outPath string

	cmd := &cobra.Command{
		Use:   "pub",
		Short: "Extract public key from private key",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyPEM, err := loadKey(keyPath)
			if err != nil {
				return err
			}
			s := cryptz25519.NewService()
			priv, err := safePrivKeyImport25519(s, keyPEM)
			if err != nil {
				return err
			}
			pub := priv.Public().(ed25519.PublicKey)
			pubPEM := s.PubKeyExport(pub)
			stdout := cmd.OutOrStdout()
			if outPath != "" {
				if err := os.WriteFile(outPath, []byte(pubPEM), 0644); err != nil {
					return fmt.Errorf("failed to write public key to %s: %w", outPath, err)
				}
				fmt.Fprintf(stdout, "Public key written to %s\n", outPath)
				return nil
			}
			fmt.Fprint(stdout, pubPEM)
			return nil
		},
	}
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "Private key file or PEM string (use '-' for stdin)")
	cmd.Flags().StringVarP(&outPath, "out", "o", "", "File to write public key to")
	cmd.MarkFlagRequired("key")

	return cmd
}

func new25519SignCmd() *cobra.Command {
	var keyPath string
	var b64 bool
	var b64url bool
	var raw bool

	cmd := &cobra.Command{
		Use:   "sign [message]",
		Short: "Sign a message using an Ed25519 private key",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyPEM, err := loadKey(keyPath)
			if err != nil {
				return err
			}
			s := cryptz25519.NewService()
			priv, err := safePrivKeyImport25519(s, keyPEM)
			if err != nil {
				return err
			}
			in, err := readInput(cmd, args)
			if err != nil {
				return err
			}
			sig := s.Sign(priv, in)
			return outputBytes(cmd, sig, b64, b64url, raw)
		},
	}
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "Private key file or PEM string (use '-' for stdin)")
	cmd.Flags().BoolVar(&b64, "b64", false, "Output signature in Base64")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Output signature in URL-safe Base64")
	cmd.Flags().BoolVarP(&raw, "raw", "r", false, "Output signature as raw binary")
	cmd.MarkFlagsMutuallyExclusive("b64", "b64url", "raw")
	cmd.MarkFlagRequired("key")

	return cmd
}

func new25519VerifyCmd() *cobra.Command {
	var keyPath string
	var sigStr string
	var sigFile string
	var b64 bool
	var b64url bool

	cmd := &cobra.Command{
		Use:   "verify [message]",
		Short: "Verify an Ed25519 signature against a message using a public key",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyPEM, err := loadKey(keyPath)
			if err != nil {
				return err
			}
			s := cryptz25519.NewService()
			pub, err := safePubKeyImport25519(s, keyPEM)
			if err != nil {
				priv, privErr := safePrivKeyImport25519(s, keyPEM)
				if privErr != nil {
					return fmt.Errorf("failed to parse key: %w", err)
				}
				pub = priv.Public().(ed25519.PublicKey)
			}

			var sigRaw []byte
			if sigFile != "" {
				fileBytes, err := os.ReadFile(sigFile)
				if err != nil {
					return fmt.Errorf("failed to read signature file: %w", err)
				}
				sigRaw = fileBytes
			} else if sigStr != "" {
				sigRaw = []byte(sigStr)
			} else {
				return fmt.Errorf("either --sig or --sig-file must be provided")
			}

			var sig []byte
			if b64url {
				blob, err := cryptzb64.UrlDecode(cleanHexInput(sigRaw))
				if err != nil {
					return fmt.Errorf("invalid b64url signature: %w", err)
				}
				sig = blob.Bytes()
			} else if b64 {
				blob, err := cryptzb64.StdDecode(cleanHexInput(sigRaw))
				if err != nil {
					return fmt.Errorf("invalid b64 signature: %w", err)
				}
				sig = blob.Bytes()
			} else {
				decoded, err := hex.DecodeString(cleanHexInput(sigRaw))
				if err != nil {
					if sigFile != "" {
						sig = sigRaw
					} else {
						return fmt.Errorf("invalid hex signature: %w", err)
					}
				} else {
					sig = decoded
				}
			}

			in, err := readInput(cmd, args)
			if err != nil {
				return err
			}

			stdout := cmd.OutOrStdout()
			ok := s.Verify(pub, in, sig)
			if !ok {
				fmt.Fprintln(stdout, "Verification failed")
				return fmt.Errorf("signature verification failed")
			}
			fmt.Fprintln(stdout, "Verification succeeded")
			return nil
		},
	}
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "Public key file or PEM string (use '-' for stdin)")
	cmd.Flags().StringVar(&sigStr, "sig", "", "Signature (hex by default, or base64)")
	cmd.Flags().StringVar(&sigFile, "sig-file", "", "File containing signature")
	cmd.Flags().BoolVar(&b64, "b64", false, "Signature is Base64 encoded")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Signature is URL-safe Base64 encoded")
	cmd.MarkFlagsMutuallyExclusive("b64", "b64url")
	cmd.MarkFlagsMutuallyExclusive("sig", "sig-file")
	cmd.MarkFlagRequired("key")

	return cmd
}

func new25519EncryptCmd() *cobra.Command {
	var keyPath string
	var label string
	var b64 bool
	var b64url bool
	var raw bool

	cmd := &cobra.Command{
		Use:   "encrypt [message]",
		Short: "Encrypt message using a Curve25519 public key (Hybrid X25519 + AES-GCM)",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyPEM, err := loadKey(keyPath)
			if err != nil {
				return err
			}
			s := cryptz25519.NewService()
			pub, err := safePubKeyImport25519(s, keyPEM)
			if err != nil {
				priv, privErr := safePrivKeyImport25519(s, keyPEM)
				if privErr != nil {
					return fmt.Errorf("failed to parse key: %w", err)
				}
				pub = priv.Public().(ed25519.PublicKey)
			}
			in, err := readInput(cmd, args)
			if err != nil {
				return err
			}

			ciphertext := s.PubEncrypt(pub, []byte(label), in)
			return outputBytes(cmd, ciphertext, b64, b64url, raw)
		},
	}
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "Public key file or PEM string (use '-' for stdin)")
	cmd.Flags().StringVar(&label, "label", "", "Optional encryption label")
	cmd.Flags().BoolVar(&b64, "b64", false, "Output ciphertext in Base64")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Output ciphertext in URL-safe Base64")
	cmd.Flags().BoolVarP(&raw, "raw", "r", false, "Output ciphertext as raw binary")
	cmd.MarkFlagsMutuallyExclusive("b64", "b64url", "raw")
	cmd.MarkFlagRequired("key")

	return cmd
}

func new25519DecryptCmd() *cobra.Command {
	var keyPath string
	var label string
	var b64 bool
	var b64url bool

	cmd := &cobra.Command{
		Use:   "decrypt [ciphertext]",
		Short: "Decrypt ciphertext using a Curve25519 private key",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyPEM, err := loadKey(keyPath)
			if err != nil {
				return err
			}
			s := cryptz25519.NewService()
			priv, err := safePrivKeyImport25519(s, keyPEM)
			if err != nil {
				return err
			}
			in, err := readInput(cmd, args)
			if err != nil {
				return err
			}

			cipherBytes, err := decodeInputBytes(in, b64, b64url)
			if err != nil {
				return err
			}

			plaintext := s.PrivDecrypt(priv, []byte(label), cipherBytes)
			fmt.Fprint(cmd.OutOrStdout(), string(plaintext))
			return nil
		},
	}
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "Private key file or PEM string (use '-' for stdin)")
	cmd.Flags().StringVar(&label, "label", "", "Optional encryption label")
	cmd.Flags().BoolVar(&b64, "b64", false, "Input ciphertext is Base64 encoded")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Input ciphertext is URL-safe Base64 encoded")
	cmd.MarkFlagsMutuallyExclusive("b64", "b64url")
	cmd.MarkFlagRequired("key")

	return cmd
}

func new25519IdCmd() *cobra.Command {
	var keyPath string

	cmd := &cobra.Command{
		Use:   "id",
		Short: "Compute key ID for public key (or private key)",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyPEM, err := loadKey(keyPath)
			if err != nil {
				return err
			}
			s := cryptz25519.NewService()
			pub, err := safePubKeyImport25519(s, keyPEM)
			if err != nil {
				priv, privErr := safePrivKeyImport25519(s, keyPEM)
				if privErr != nil {
					return fmt.Errorf("failed to parse key: %w", err)
				}
				pub = priv.Public().(ed25519.PublicKey)
			}
			id := s.PubKeyId(pub)
			fmt.Fprintln(cmd.OutOrStdout(), id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "Key file or PEM string (use '-' for stdin)")
	cmd.MarkFlagRequired("key")

	return cmd
}

func safePrivKeyImport25519(s cryptz25519service.Service25519, keyPEM string) (priv ed25519.PrivateKey, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to import private key: %v", r)
		}
	}()
	return s.PrivKeyImport(keyPEM), nil
}

func safePubKeyImport25519(s cryptz25519service.Service25519, keyPEM string) (pub ed25519.PublicKey, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to import public key: %v", r)
		}
	}()
	return s.PubKeyImport(keyPEM), nil
}
