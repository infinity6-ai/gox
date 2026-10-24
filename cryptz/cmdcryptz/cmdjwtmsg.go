package cmdcryptz

import (
	"fmt"

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
		Use:     "encrypt [message]",
		Aliases: []string{"send"},
		Short:   "Encrypt and sign a message from sender to recipient using KeyRing",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadKeyRingFile(krPath)
			if err != nil {
				return err
			}

			in, err := readInput(args)
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
	var b64 bool
	var b64url bool

	cmd := &cobra.Command{
		Use:     "decrypt [ciphertext]",
		Aliases: []string{"receive"},
		Short:   "Decrypt and verify a message using recipient private key and sender public key from KeyRing",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadKeyRingFile(krPath)
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

			mkr := cryptzjwtmsg.FromKeyRing(kr)
			dec := cryptzjwtmsg.NewDecryptor()
			dec.KeysLoader = mkr.DecryptorKeysLoader()
			dec.CipheredData = cipherBytes

			plain, err := safeMsgDecrypt(dec)
			if err != nil {
				fmt.Fprintln(cmd.OutOrStderr(), "Decryption failed:", err)
				return fmt.Errorf("decryption failed: %w", err)
			}

			fmt.Fprint(cmd.OutOrStdout(), string(plain.Data))
			return nil
		},
	}

	cmd.Flags().StringVarP(&krPath, "kr", "k", "keyring.json", "Keyring file path")
	cmd.Flags().BoolVar(&b64, "b64", false, "Input ciphertext is Base64")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Input ciphertext is URL-safe Base64")
	cmd.MarkFlagsMutuallyExclusive("b64", "b64url")

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
