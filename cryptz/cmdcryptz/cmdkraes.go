package cmdcryptz

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/infinity6-ai/gox/cryptz/cryptzaes"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt/cryptzjwtstate"
	"github.com/spf13/cobra"
)

func newKrAesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kr-aes",
		Aliases: []string{"keyring-aes", "aes-kr"},
		Short:   "Symmetric AES KeyRing management (init, gen, gen-pair, add, add-pair, del, del-pair, exp-key, exp-pair, export, list)",
	}

	cmd.AddCommand(
		newKrAesInitCmd(),
		newKrAesGenCmd(),
		newKrAesGenPairCmd(),
		newKrAesAddCmd(),
		newKrAesAddPairCmd(),
		newKrAesDelCmd(),
		newKrAesDelPairCmd(),
		newKrAesExpKeyCmd(),
		newKrAesExpPairCmd(),
		newKrAesExportCmd(),
		newKrAesListCmd(),
	)

	return cmd
}

func saveStateKeyRingFile(filePath string, kr *cryptzjwtstate.KeyRing) error {
	if filePath == "" {
		filePath = "state_keyring.json"
	}
	data, err := kr.Export()
	if err != nil {
		return fmt.Errorf("failed to export state keyring: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write state keyring to %s: %w", filePath, err)
	}
	return nil
}

func newKrAesInitCmd() *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new empty symmetric AES keyring file",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr := cryptzjwtstate.NewKeyRing()
			if err := saveStateKeyRingFile(filePath, kr); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Initialized empty AES keyring at %s\n", filePath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Keyring file path")
	cmd.MarkFlagRequired("file")
	return cmd
}

func newKrAesGenCmd() *cobra.Command {
	var filePath string
	var owner string
	var size int

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate a new AES key for an owner in AES keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			var kr *cryptzjwtstate.KeyRing
			if _, err := os.Stat(filePath); err == nil {
				kr, err = loadStateKeyRingFile(filePath)
				if err != nil {
					return err
				}
			} else {
				kr = cryptzjwtstate.NewKeyRing()
			}

			key := kr.Generate(owner, size)
			if err := saveStateKeyRingFile(filePath, kr); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Generated AES key for owner %s (%d bytes) in %s: %s\n", owner, len(key), filePath, hex.EncodeToString(key))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "state_keyring.json", "Keyring file path")
	cmd.Flags().StringVarP(&owner, "owner", "o", "", "Owner name")
	cmd.Flags().IntVarP(&size, "size", "s", 32, "Key size in bytes (16, 24, or 32)")
	cmd.MarkFlagRequired("owner")
	return cmd
}

func newKrAesGenPairCmd() *cobra.Command {
	var filePath string
	var src string
	var dst string
	var size int

	cmd := &cobra.Command{
		Use:   "gen-pair",
		Short: "Generate a pairwise AES key for src and dst in AES keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			var kr *cryptzjwtstate.KeyRing
			if _, err := os.Stat(filePath); err == nil {
				kr, err = loadStateKeyRingFile(filePath)
				if err != nil {
					return err
				}
			} else {
				kr = cryptzjwtstate.NewKeyRing()
			}

			key := kr.GeneratePair(src, dst, size)
			if err := saveStateKeyRingFile(filePath, kr); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Generated pairwise AES key for %s <-> %s in %s: %s\n", src, dst, filePath, hex.EncodeToString(key))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "state_keyring.json", "Keyring file path")
	cmd.Flags().StringVar(&src, "src", "", "Source owner name")
	cmd.Flags().StringVar(&dst, "dst", "", "Destination owner name")
	cmd.Flags().IntVarP(&size, "size", "s", 32, "Key size in bytes (16, 24, or 32)")
	cmd.MarkFlagRequired("src")
	cmd.MarkFlagRequired("dst")
	return cmd
}

func newKrAesAddCmd() *cobra.Command {
	var filePath string
	var owner string
	var keyHex string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an existing AES key in hex for an owner in AES keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadStateKeyRingFile(filePath)
			if err != nil {
				return err
			}

			keyBytes, err := hex.DecodeString(cleanHexInput([]byte(keyHex)))
			if err != nil {
				return fmt.Errorf("invalid key hex: %w", err)
			}

			kr.Add(owner, cryptzaes.Key(keyBytes))
			if err := saveStateKeyRingFile(filePath, kr); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Added AES key for owner %s in %s\n", owner, filePath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "state_keyring.json", "Keyring file path")
	cmd.Flags().StringVarP(&owner, "owner", "o", "", "Owner name")
	cmd.Flags().StringVarP(&keyHex, "key", "k", "", "AES key in hex")
	cmd.MarkFlagRequired("owner")
	cmd.MarkFlagRequired("key")
	return cmd
}

func newKrAesAddPairCmd() *cobra.Command {
	var filePath string
	var src string
	var dst string
	var keyHex string

	cmd := &cobra.Command{
		Use:   "add-pair",
		Short: "Add an existing pairwise AES key in hex in AES keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadStateKeyRingFile(filePath)
			if err != nil {
				return err
			}

			keyBytes, err := hex.DecodeString(cleanHexInput([]byte(keyHex)))
			if err != nil {
				return fmt.Errorf("invalid key hex: %w", err)
			}

			kr.AddPair(src, dst, cryptzaes.Key(keyBytes))
			if err := saveStateKeyRingFile(filePath, kr); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Added pairwise AES key for %s <-> %s in %s\n", src, dst, filePath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "state_keyring.json", "Keyring file path")
	cmd.Flags().StringVar(&src, "src", "", "Source owner name")
	cmd.Flags().StringVar(&dst, "dst", "", "Destination owner name")
	cmd.Flags().StringVarP(&keyHex, "key", "k", "", "AES key in hex")
	cmd.MarkFlagRequired("src")
	cmd.MarkFlagRequired("dst")
	cmd.MarkFlagRequired("key")
	return cmd
}

func newKrAesDelCmd() *cobra.Command {
	var filePath string
	var owner string

	cmd := &cobra.Command{
		Use:     "del",
		Aliases: []string{"rm", "delete"},
		Short:   "Delete an owner key from AES keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadStateKeyRingFile(filePath)
			if err != nil {
				return err
			}

			kr.Delete(owner)
			if err := saveStateKeyRingFile(filePath, kr); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted AES key for owner %s from %s\n", owner, filePath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "state_keyring.json", "Keyring file path")
	cmd.Flags().StringVarP(&owner, "owner", "o", "", "Owner name to delete")
	cmd.MarkFlagRequired("owner")
	return cmd
}

func newKrAesDelPairCmd() *cobra.Command {
	var filePath string
	var src string
	var dst string

	cmd := &cobra.Command{
		Use:     "del-pair",
		Aliases: []string{"rm-pair", "delete-pair"},
		Short:   "Delete a pairwise key from AES keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadStateKeyRingFile(filePath)
			if err != nil {
				return err
			}

			kr.DeletePair(src, dst)
			if err := saveStateKeyRingFile(filePath, kr); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted pairwise AES key for %s <-> %s from %s\n", src, dst, filePath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "state_keyring.json", "Keyring file path")
	cmd.Flags().StringVar(&src, "src", "", "Source owner name")
	cmd.Flags().StringVar(&dst, "dst", "", "Destination owner name")
	cmd.MarkFlagRequired("src")
	cmd.MarkFlagRequired("dst")
	return cmd
}

func newKrAesExpKeyCmd() *cobra.Command {
	var filePath string
	var owner string
	var outFile string

	cmd := &cobra.Command{
		Use:   "exp-key",
		Short: "Export AES key in hex for an owner",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadStateKeyRingFile(filePath)
			if err != nil {
				return err
			}

			key := kr.GetRawKey(owner)
			if len(key) == 0 {
				return fmt.Errorf("no AES key found for owner %s", owner)
			}

			hexKey := hex.EncodeToString(key)
			if outFile != "" {
				if err := os.WriteFile(outFile, []byte(hexKey), 0600); err != nil {
					return fmt.Errorf("failed to write key to %s: %w", outFile, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Exported key to %s\n", outFile)
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), hexKey)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "state_keyring.json", "Keyring file path")
	cmd.Flags().StringVarP(&owner, "owner", "o", "", "Owner name")
	cmd.Flags().StringVar(&outFile, "out", "", "File path to write key")
	cmd.MarkFlagRequired("owner")
	return cmd
}

func newKrAesExpPairCmd() *cobra.Command {
	var filePath string
	var src string
	var dst string
	var outFile string

	cmd := &cobra.Command{
		Use:   "exp-pair",
		Short: "Export pairwise AES key in hex for src and dst",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadStateKeyRingFile(filePath)
			if err != nil {
				return err
			}

			key := kr.GetKey(src, dst)
			if len(key) == 0 {
				return fmt.Errorf("no AES key found for pair %s <-> %s", src, dst)
			}

			hexKey := hex.EncodeToString(key)
			if outFile != "" {
				if err := os.WriteFile(outFile, []byte(hexKey), 0600); err != nil {
					return fmt.Errorf("failed to write key to %s: %w", outFile, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Exported key to %s\n", outFile)
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), hexKey)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "state_keyring.json", "Keyring file path")
	cmd.Flags().StringVar(&src, "src", "", "Source owner name")
	cmd.Flags().StringVar(&dst, "dst", "", "Destination owner name")
	cmd.Flags().StringVar(&outFile, "out", "", "File path to write key")
	cmd.MarkFlagRequired("src")
	cmd.MarkFlagRequired("dst")
	return cmd
}

func newKrAesExportCmd() *cobra.Command {
	var filePath string
	var outFile string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export entire AES keyring JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadStateKeyRingFile(filePath)
			if err != nil {
				return err
			}

			data, err := kr.Export()
			if err != nil {
				return err
			}

			if outFile != "" {
				if err := os.WriteFile(outFile, data, 0600); err != nil {
					return fmt.Errorf("failed to write export to %s: %w", outFile, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Exported AES keyring to %s\n", outFile)
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "state_keyring.json", "Keyring file path")
	cmd.Flags().StringVarP(&outFile, "out", "o", "", "File path to write export")
	return cmd
}

func newKrAesListCmd() *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List owners and pairs in AES keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadStateKeyRingFile(filePath)
			if err != nil {
				return err
			}

			names := kr.GetNames()
			if len(names) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Keyring is empty")
				return nil
			}

			for _, name := range names {
				key := kr.GetRawKey(name)
				fmt.Fprintf(cmd.OutOrStdout(), "Key: %s (%d bytes)\n", name, len(key))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "state_keyring.json", "Keyring file path")
	return cmd
}
