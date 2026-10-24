package cmdcryptz

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/infinity6-ai/gox/cryptz/cryptzsalt"
	"github.com/spf13/cobra"
)

func newSaltCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "salt",
		Short: "Salt generate/verify",
	}

	genCmd := &cobra.Command{
		Use:   "gen [text]",
		Short: "Generate salted hash",
		RunE: func(cmd *cobra.Command, args []string) error {
			in, err := readInput(args)
			if err != nil {
				return err
			}
			res, err := cryptzsalt.Generate(in)
			if err != nil {
				return err
			}
			fmt.Println(hex.EncodeToString(res.Format()))
			return nil
		},
	}

	var original string
	verifyCmd := &cobra.Command{
		Use:   "verify [bundle_hex]",
		Short: "Verify a salted hash",
		RunE: func(cmd *cobra.Command, args []string) error {
			in, err := readInput(args)
			if err != nil {
				return err
			}
			bundleBytes, err := hex.DecodeString(cleanHexInput(in))
			if err != nil {
				return fmt.Errorf("invalid bundle hex: %w", err)
			}
			_, err = cryptzsalt.Verify(bundleBytes, original)
			if err != nil {
				fmt.Println("Verification failed:", err)
				os.Exit(1)
			}
			fmt.Println("Verification succeeded")
			return nil
		},
	}
	verifyCmd.Flags().StringVarP(&original, "original", "o", "", "Original text to verify against")
	verifyCmd.MarkFlagRequired("original")

	cmd.AddCommand(genCmd, verifyCmd)
	return cmd
}
