package cmdcryptz

import (
	"fmt"
	"os"

	"github.com/infinity6-ai/gox/cryptz/cryptzrsa"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsakeyring"
	"github.com/spf13/cobra"
)

func newKeyringCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "keyring",
		Aliases: []string{"kr"},
		Short:   "RSA KeyRing management (init, gen, add-priv, add-pub, export, list)",
	}

	cmd.AddCommand(
		newKeyringInitCmd(),
		newKeyringGenCmd(),
		newKeyringAddPrivCmd(),
		newKeyringAddPubCmd(),
		newKeyringExportCmd(),
		newKeyringListCmd(),
		newKeyringExpPrivCmd(),
		newKeyringExpPubCmd(),
		newKeyringDelKeyCmd(),
		newKeyringDelPrivCmd(),
	)

	return cmd
}

func loadKeyRingFile(filePath string) (*cryptzrsakeyring.KeyRing, error) {
	if filePath == "" {
		filePath = "keyring.json"
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keyring file %s: %w", filePath, err)
	}
	kr, err := cryptzrsakeyring.Import(data)
	if err != nil {
		return nil, fmt.Errorf("failed to import keyring from %s: %w", filePath, err)
	}
	return kr, nil
}

func saveKeyRingFile(filePath string, kr *cryptzrsakeyring.KeyRing) error {
	if filePath == "" {
		filePath = "keyring.json"
	}
	data, err := kr.ExportAll()
	if err != nil {
		return fmt.Errorf("failed to export keyring: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write keyring to %s: %w", filePath, err)
	}
	return nil
}

func newKeyringInitCmd() *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new empty keyring file",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr := cryptzrsakeyring.NewKeyRing()
			if err := saveKeyRingFile(filePath, kr); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Initialized empty keyring at %s\n", filePath)
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Keyring file path")
	cmd.MarkFlagRequired("file")
	return cmd
}

func newKeyringGenCmd() *cobra.Command {
	var filePath string
	var owner string
	var bits int

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate a new RSA key for an owner and add to keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			var kr *cryptzrsakeyring.KeyRing
			if _, err := os.Stat(filePath); err == nil {
				kr, err = loadKeyRingFile(filePath)
				if err != nil {
					return err
				}
			} else {
				kr = cryptzrsakeyring.NewKeyRing()
			}

			kr.Generate(owner, bits)
			if err := saveKeyRingFile(filePath, kr); err != nil {
				return err
			}

			keyId := kr.Current(owner).KeyId()
			fmt.Fprintf(cmd.OutOrStdout(), "Generated key %s for owner %s in %s\n", keyId, owner, filePath)
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "keyring.json", "Keyring file path")
	cmd.Flags().StringVarP(&owner, "owner", "o", "", "Owner name (must match ^[a-z][a-z0-9]*$)")
	cmd.Flags().IntVarP(&bits, "bits", "b", 2048, "Key size in bits")
	cmd.MarkFlagRequired("owner")
	return cmd
}

func newKeyringAddPrivCmd() *cobra.Command {
	var filePath string
	var owner string
	var keyPath string

	cmd := &cobra.Command{
		Use:   "add-priv",
		Short: "Add an existing private key to an owner in keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadKeyRingFile(filePath)
			if err != nil {
				return err
			}

			keyPEM, err := loadKey(keyPath)
			if err != nil {
				return err
			}

			s := cryptzrsa.NewService()
			priv, err := safePrivKeyImport(s, keyPEM)
			if err != nil {
				return err
			}

			kr.AddPrivate(owner, priv)
			if err := saveKeyRingFile(filePath, kr); err != nil {
				return err
			}

			keyId := kr.Current(owner).KeyId()
			fmt.Fprintf(cmd.OutOrStdout(), "Added private key %s for owner %s in %s\n", keyId, owner, filePath)
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "keyring.json", "Keyring file path")
	cmd.Flags().StringVarP(&owner, "owner", "o", "", "Owner name")
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "Private key file or PEM string")
	cmd.MarkFlagRequired("owner")
	cmd.MarkFlagRequired("key")
	return cmd
}

func newKeyringAddPubCmd() *cobra.Command {
	var filePath string
	var owner string
	var keyPath string

	cmd := &cobra.Command{
		Use:   "add-pub",
		Short: "Add an existing public key to an owner in keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadKeyRingFile(filePath)
			if err != nil {
				return err
			}

			keyPEM, err := loadKey(keyPath)
			if err != nil {
				return err
			}

			s := cryptzrsa.NewService()
			pub, err := safePubKeyImport(s, keyPEM)
			if err != nil {
				return err
			}

			added := kr.AddPublic(owner, pub)
			if err := saveKeyRingFile(filePath, kr); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Added public key %s for owner %s in %s\n", added.KeyId(), owner, filePath)
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "keyring.json", "Keyring file path")
	cmd.Flags().StringVarP(&owner, "owner", "o", "", "Owner name")
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "Public key file or PEM string")
	cmd.MarkFlagRequired("owner")
	cmd.MarkFlagRequired("key")
	return cmd
}

func newKeyringExportCmd() *cobra.Command {
	var filePath string
	var pubsOnly bool
	var outFile string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export keyring to stdout or file",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadKeyRingFile(filePath)
			if err != nil {
				return err
			}

			var data []byte
			if pubsOnly {
				data, err = kr.ExportPubs()
			} else {
				data, err = kr.ExportAll()
			}
			if err != nil {
				return err
			}

			if outFile != "" {
				if err := os.WriteFile(outFile, data, 0644); err != nil {
					return fmt.Errorf("failed to write export to %s: %w", outFile, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Exported keyring to %s\n", outFile)
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "keyring.json", "Keyring file path")
	cmd.Flags().BoolVar(&pubsOnly, "pubs-only", false, "Export only public keys (strip private keys)")
	cmd.Flags().StringVarP(&outFile, "out", "o", "", "File path to write export")
	return cmd
}

func newKeyringListCmd() *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List owners and keys in keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadKeyRingFile(filePath)
			if err != nil {
				return err
			}

			owners := kr.GetOwnerNames()
			if len(owners) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Keyring is empty")
				return nil
			}

			for _, ownerName := range owners {
				currentKey := kr.Current(ownerName)
				keys := kr.Owners(ownerName)
				fmt.Fprintf(cmd.OutOrStdout(), "Owner: %s (current: %s)\n", ownerName, currentKey.KeyId())
				_ = keys
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "keyring.json", "Keyring file path")
	return cmd
}

func getKeyForOwner(kr *cryptzrsakeyring.KeyRing, owner, keyId string) (*cryptzrsakeyring.Key, error) {
	if keyId != "" {
		k := kr.KeyIdOrNil(keyId)
		if k == nil {
			return nil, fmt.Errorf("key with id %s not found in keyring", keyId)
		}
		if k.Owner() != owner {
			return nil, fmt.Errorf("key %s belongs to owner %s, not %s", keyId, k.Owner(), owner)
		}
		return k, nil
	}

	var currentKey *cryptzrsakeyring.Key
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("owner %s not found in keyring: %v", owner, r)
			}
		}()
		currentKey = kr.Current(owner)
	}()
	if err != nil {
		return nil, err
	}
	return currentKey, nil
}

func newKeyringExpPrivCmd() *cobra.Command {
	var filePath string
	var owner string
	var keyId string
	var outFile string

	cmd := &cobra.Command{
		Use:     "exp-priv",
		Aliases: []string{"export-priv", "priv"},
		Short:   "Export private key PEM for an owner",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadKeyRingFile(filePath)
			if err != nil {
				return err
			}

			k, err := getKeyForOwner(kr, owner, keyId)
			if err != nil {
				return err
			}

			if k.Private() == nil {
				return fmt.Errorf("key %s for owner %s does not contain a private key", k.KeyId(), owner)
			}

			s := cryptzrsa.NewService()
			pemStr := s.PrivKeyExport(k.Private())

			if outFile != "" {
				if err := os.WriteFile(outFile, []byte(pemStr), 0600); err != nil {
					return fmt.Errorf("failed to write private key to %s: %w", outFile, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Exported private key to %s\n", outFile)
				return nil
			}

			fmt.Fprint(cmd.OutOrStdout(), pemStr)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "keyring.json", "Keyring file path")
	cmd.Flags().StringVarP(&owner, "owner", "o", "", "Owner name")
	cmd.Flags().StringVar(&keyId, "key-id", "", "Specific key ID (defaults to owner's current key)")
	cmd.Flags().StringVar(&outFile, "out", "", "File path to write PEM private key")
	cmd.MarkFlagRequired("owner")

	return cmd
}

func newKeyringExpPubCmd() *cobra.Command {
	var filePath string
	var owner string
	var keyId string
	var outFile string

	cmd := &cobra.Command{
		Use:     "exp-pub",
		Aliases: []string{"export-pub", "pub"},
		Short:   "Export public key PEM for an owner",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadKeyRingFile(filePath)
			if err != nil {
				return err
			}

			k, err := getKeyForOwner(kr, owner, keyId)
			if err != nil {
				return err
			}

			s := cryptzrsa.NewService()
			pemStr := s.PubKeyExport(k.Public())

			if outFile != "" {
				if err := os.WriteFile(outFile, []byte(pemStr), 0644); err != nil {
					return fmt.Errorf("failed to write public key to %s: %w", outFile, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Exported public key to %s\n", outFile)
				return nil
			}

			fmt.Fprint(cmd.OutOrStdout(), pemStr)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "keyring.json", "Keyring file path")
	cmd.Flags().StringVarP(&owner, "owner", "o", "", "Owner name")
	cmd.Flags().StringVar(&keyId, "key-id", "", "Specific key ID (defaults to owner's current key)")
	cmd.Flags().StringVar(&outFile, "out", "", "File path to write PEM public key")
	cmd.MarkFlagRequired("owner")

	return cmd
}

func newKeyringDelKeyCmd() *cobra.Command {
	var filePath string
	var owner string
	var keyId string

	cmd := &cobra.Command{
		Use:     "del-key",
		Aliases: []string{"rm-key", "delete-key"},
		Short:   "Delete a key pair from an owner in keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadKeyRingFile(filePath)
			if err != nil {
				return err
			}

			if keyId == "" {
				k, err := getKeyForOwner(kr, owner, "")
				if err != nil {
					return err
				}
				keyId = k.KeyId()
			}

			if err := kr.DeleteKey(owner, keyId); err != nil {
				return err
			}

			if err := saveKeyRingFile(filePath, kr); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted key %s for owner %s from %s\n", keyId, owner, filePath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "keyring.json", "Keyring file path")
	cmd.Flags().StringVarP(&owner, "owner", "o", "", "Owner name")
	cmd.Flags().StringVar(&keyId, "key-id", "", "Specific key ID to delete (defaults to owner's current key)")
	cmd.MarkFlagRequired("owner")

	return cmd
}

func newKeyringDelPrivCmd() *cobra.Command {
	var filePath string
	var owner string
	var keyId string

	cmd := &cobra.Command{
		Use:     "del-priv",
		Aliases: []string{"rm-priv", "delete-priv"},
		Short:   "Delete only the private key for an owner's key (retains public key)",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr, err := loadKeyRingFile(filePath)
			if err != nil {
				return err
			}

			if keyId == "" {
				k, err := getKeyForOwner(kr, owner, "")
				if err != nil {
					return err
				}
				keyId = k.KeyId()
			}

			if err := kr.DeletePrivate(owner, keyId); err != nil {
				return err
			}

			if err := saveKeyRingFile(filePath, kr); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted private key for key %s (owner %s) in %s\n", keyId, owner, filePath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "keyring.json", "Keyring file path")
	cmd.Flags().StringVarP(&owner, "owner", "o", "", "Owner name")
	cmd.Flags().StringVar(&keyId, "key-id", "", "Specific key ID (defaults to owner's current key)")
	cmd.MarkFlagRequired("owner")

	return cmd
}
