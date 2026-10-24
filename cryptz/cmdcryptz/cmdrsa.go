package cmdcryptz

import (
	"fmt"
	"os"

	"github.com/infinity6-ai/gox/cryptz/cryptzrsa"
	"github.com/spf13/cobra"
)

func newRsaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rsa",
		Short: "RSA utilities",
	}

	var bits int
	var privOut string
	var pubOut string

	genCmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate RSA key pair",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := cryptzrsa.NewService()
			priv := s.PrivKeyCreate(bits)
			privStr := s.PrivKeyExport(priv)
			pubStr := s.PubKeyExport(&priv.PublicKey)

			if privOut != "" {
				if err := os.WriteFile(privOut, []byte(privStr), 0600); err != nil {
					return err
				}
				fmt.Printf("Private key written to %s\n", privOut)
			} else {
				fmt.Println(privStr)
			}

			if pubOut != "" {
				if err := os.WriteFile(pubOut, []byte(pubStr), 0644); err != nil {
					return err
				}
				fmt.Printf("Public key written to %s\n", pubOut)
			} else {
				fmt.Println(pubStr)
			}
			return nil
		},
	}
	genCmd.Flags().IntVarP(&bits, "bits", "b", 2048, "Key size in bits")
	genCmd.Flags().StringVar(&privOut, "priv-out", "", "File to write private key to")
	genCmd.Flags().StringVar(&pubOut, "pub-out", "", "File to write public key to")

	cmd.AddCommand(genCmd)
	return cmd
}
