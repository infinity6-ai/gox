package cmdcryptz

import (
	"encoding/hex"
	"fmt"

	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
	"github.com/infinity6-ai/gox/cryptz/cryptzrand"
	"github.com/spf13/cobra"
)

func newRandCmd() *cobra.Command {
	var size int
	var b64 bool
	var b64url bool
	var raw bool

	cmd := &cobra.Command{
		Use:   "rand",
		Short: "Generate random bytes",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cryptzrand.Rand(size)
			stdout := cmd.OutOrStdout()
			if raw {
				_, err := stdout.Write(out)
				return err
			}
			if b64url {
				fmt.Fprintln(stdout, cryptzb64.UrlEncode(out).String())
				return nil
			}
			if b64 {
				fmt.Fprintln(stdout, cryptzb64.StdEncode(out).String())
				return nil
			}
			fmt.Fprintln(stdout, hex.EncodeToString(out))
			return nil
		},
	}

	cmd.Flags().IntVarP(&size, "size", "s", 32, "Number of bytes to generate")
	cmd.Flags().BoolVar(&b64, "b64", false, "Output in Base64 instead of Hex")
	cmd.Flags().BoolVar(&b64url, "b64url", false, "Output in URL-safe Base64 instead of Hex")
	cmd.Flags().BoolVarP(&raw, "raw", "r", false, "Output raw binary data directly to stdout")

	cmd.MarkFlagsMutuallyExclusive("b64", "b64url", "raw")

	return cmd
}
