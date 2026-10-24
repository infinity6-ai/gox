package cmdcryptz

import (
	"encoding/hex"
	"fmt"

	"github.com/infinity6-ai/gox/cryptz/cryptzhash"
	"github.com/spf13/cobra"
)

func newHashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hash",
		Short: "Hashing utilities",
	}

	sha256Cmd := &cobra.Command{
		Use:   "sha256 [text]",
		Short: "SHA256 hash",
		RunE: func(cmd *cobra.Command, args []string) error {
			in, err := readInput(args)
			if err != nil {
				return err
			}
			out, err := cryptzhash.SHA256Data(in)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), hex.EncodeToString(out))
			return nil
		},
	}

	md5Cmd := &cobra.Command{
		Use:   "md5 [text]",
		Short: "MD5 hash",
		RunE: func(cmd *cobra.Command, args []string) error {
			in, err := readInput(args)
			if err != nil {
				return err
			}
			out, err := cryptzhash.MD5Data(in)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), hex.EncodeToString(out))
			return nil
		},
	}

	cmd.AddCommand(sha256Cmd, md5Cmd)
	return cmd
}
