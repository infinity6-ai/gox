package cmdcryptz

import (
	"encoding/hex"
	"fmt"

	"github.com/infinity6-ai/gox/cryptz/cryptzaes"
	"github.com/spf13/cobra"
)

func newAesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aes",
		Short: "AES GCM encrypt/decrypt",
	}

	var keyHex string

	encryptCmd := &cobra.Command{
		Use:   "encrypt [text]",
		Short: "Encrypt using AES GCM",
		RunE: func(cmd *cobra.Command, args []string) error {
			in, err := readInput(args)
			if err != nil {
				return err
			}
			keyBytes, err := hex.DecodeString(keyHex)
			if err != nil {
				return fmt.Errorf("invalid key hex: %w", err)
			}
			out, err := cryptzaes.AESCrypt(cryptzaes.Key(keyBytes), in)
			if err != nil {
				return err
			}
			fmt.Print(hex.EncodeToString(out))
			return nil
		},
	}
	encryptCmd.Flags().StringVarP(&keyHex, "key", "k", "", "AES Key in hex (16, 24, 32 bytes)")
	encryptCmd.MarkFlagRequired("key")

	decryptCmd := &cobra.Command{
		Use:   "decrypt [hex_ciphertext]",
		Short: "Decrypt using AES GCM",
		RunE: func(cmd *cobra.Command, args []string) error {
			in, err := readInput(args)
			if err != nil {
				return err
			}
			cipherBytes, err := hex.DecodeString(cleanHexInput(in))
			if err != nil {
				return fmt.Errorf("invalid ciphertext hex: %w", err)
			}
			keyBytes, err := hex.DecodeString(keyHex)
			if err != nil {
				return fmt.Errorf("invalid key hex: %w", err)
			}
			out, err := cryptzaes.AESDecrypt(cryptzaes.Key(keyBytes), cipherBytes)
			if err != nil {
				return err
			}
			fmt.Print(string(out.Bytes()))
			return nil
		},
	}
	decryptCmd.Flags().StringVarP(&keyHex, "key", "k", "", "AES Key in hex")
	decryptCmd.MarkFlagRequired("key")

	cmd.AddCommand(encryptCmd, decryptCmd)
	return cmd
}
