package cmdcryptz

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa"
	"github.com/spf13/cobra"
)

func newRsaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rsa",
		Short: "RSA utilities (keygen, pubkey extraction, sign, verify, encrypt, decrypt, id)",
	}

	cmd.AddCommand(
		newRsaGenCmd(),
		newRsaPubCmd(),
		newRsaSignCmd(),
		newRsaVerifyCmd(),
		newRsaEncryptCmd(),
		newRsaDecryptCmd(),
		newRsaIdCmd(),
	)

	return cmd
}

func newRsaGenCmd() *cobra.Command {
	var bits int
	var privOut string
	var pubOut string

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate RSA key pair",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := cryptzrsa.NewService()
			priv := s.PrivKeyCreate(bits)
			privStr := s.PrivKeyExport(priv)
			pubStr := s.PubKeyExport(&priv.PublicKey)
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
	cmd.Flags().IntVarP(&bits, "bits", "b", 2048, "Key size in bits")
	cmd.Flags().StringVar(&privOut, "priv-out", "", "File to write private key to")
	cmd.Flags().StringVar(&pubOut, "pub-out", "", "File to write public key to")

	return cmd
}

func newRsaPubCmd() *cobra.Command {
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
			s := cryptzrsa.NewService()
			priv, err := safePrivKeyImport(s, keyPEM)
			if err != nil {
				return err
			}
			pubPEM := s.PubKeyExport(&priv.PublicKey)
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

func newRsaSignCmd() *cobra.Command {
	var keyPath string
	var b64 bool
	var b64url bool
	var raw bool

	cmd := &cobra.Command{
		Use:   "sign [message]",
		Short: "Sign a message using an RSA private key (SHA-256 with PKCS#1 v1.5)",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyPEM, err := loadKey(keyPath)
			if err != nil {
				return err
			}
			s := cryptzrsa.NewService()
			priv, err := safePrivKeyImport(s, keyPEM)
			if err != nil {
				return err
			}
			in, err := readInput(args)
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

func newRsaVerifyCmd() *cobra.Command {
	var keyPath string
	var sigStr string
	var sigFile string
	var b64 bool
	var b64url bool

	cmd := &cobra.Command{
		Use:   "verify [message]",
		Short: "Verify an RSA signature against a message using a public key",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyPEM, err := loadKey(keyPath)
			if err != nil {
				return err
			}
			s := cryptzrsa.NewService()
			pub, err := safePubKeyImport(s, keyPEM)
			if err != nil {
				// Also allow passing a private key where we extract the public key
				priv, privErr := safePrivKeyImport(s, keyPEM)
				if privErr != nil {
					return fmt.Errorf("failed to parse key: %w", err)
				}
				pub = &priv.PublicKey
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

			in, err := readInput(args)
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

func newRsaEncryptCmd() *cobra.Command {
	var keyPath string
	var label string
	var b64 bool
	var b64url bool
	var raw bool

	cmd := &cobra.Command{
		Use:   "encrypt [message]",
		Short: "Encrypt message using an RSA public key with OAEP (SHA-256)",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyPEM, err := loadKey(keyPath)
			if err != nil {
				return err
			}
			s := cryptzrsa.NewService()
			pub, err := safePubKeyImport(s, keyPEM)
			if err != nil {
				priv, privErr := safePrivKeyImport(s, keyPEM)
				if privErr != nil {
					return fmt.Errorf("failed to parse key: %w", err)
				}
				pub = &priv.PublicKey
			}
			in, err := readInput(args)
			if err != nil {
				return err
			}

			ciphertext := s.PubEncryptOEAP(pub, []byte(label), in)
			return outputBytes(cmd, ciphertext, b64, b64url, raw)
		},
	}
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "Public key file or PEM string (use '-' for stdin)")
	cmd.Flags().StringVar(&label, "label", "", "Optional OAEP label")
	cmd.Flags().BoolVar(&b64, "b64", false, "Output ciphertext in Base64")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Output ciphertext in URL-safe Base64")
	cmd.Flags().BoolVarP(&raw, "raw", "r", false, "Output ciphertext as raw binary")
	cmd.MarkFlagsMutuallyExclusive("b64", "b64url", "raw")
	cmd.MarkFlagRequired("key")

	return cmd
}

func newRsaDecryptCmd() *cobra.Command {
	var keyPath string
	var label string
	var b64 bool
	var b64url bool

	cmd := &cobra.Command{
		Use:   "decrypt [ciphertext]",
		Short: "Decrypt ciphertext using an RSA private key with OAEP (SHA-256)",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyPEM, err := loadKey(keyPath)
			if err != nil {
				return err
			}
			s := cryptzrsa.NewService()
			priv, err := safePrivKeyImport(s, keyPEM)
			if err != nil {
				return err
			}
			in, err := readInput(args)
			if err != nil {
				return err
			}

			cipherBytes, err := decodeInputBytes(in, b64, b64url)
			if err != nil {
				return err
			}

			plaintext := s.PrivDecryptOEAP(priv, []byte(label), cipherBytes)
			fmt.Fprint(cmd.OutOrStdout(), string(plaintext))
			return nil
		},
	}
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "Private key file or PEM string (use '-' for stdin)")
	cmd.Flags().StringVar(&label, "label", "", "Optional OAEP label")
	cmd.Flags().BoolVar(&b64, "b64", false, "Input ciphertext is Base64 encoded")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Input ciphertext is URL-safe Base64 encoded")
	cmd.MarkFlagsMutuallyExclusive("b64", "b64url")
	cmd.MarkFlagRequired("key")

	return cmd
}

func newRsaIdCmd() *cobra.Command {
	var keyPath string

	cmd := &cobra.Command{
		Use:   "id",
		Short: "Compute key ID for public key (or private key)",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyPEM, err := loadKey(keyPath)
			if err != nil {
				return err
			}
			s := cryptzrsa.NewService()
			pub, err := safePubKeyImport(s, keyPEM)
			if err != nil {
				priv, privErr := safePrivKeyImport(s, keyPEM)
				if privErr != nil {
					return fmt.Errorf("failed to parse key: %w", err)
				}
				pub = &priv.PublicKey
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
