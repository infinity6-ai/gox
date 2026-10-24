package cmdcryptz

import (
	"fmt"

	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
	"github.com/spf13/cobra"
)

func newB64Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "b64",
		Short: "Base64 encode/decode",
	}

	var urlSafe bool

	encodeCmd := &cobra.Command{
		Use:   "encode [text]",
		Short: "Encode to Base64",
		RunE: func(cmd *cobra.Command, args []string) error {
			in, err := readInput(args)
			if err != nil {
				return err
			}
			var out []byte
			if urlSafe {
				out = cryptzb64.UrlEncode(in).Bytes()
			} else {
				out = cryptzb64.StdEncode(in).Bytes()
			}
			fmt.Fprint(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
	encodeCmd.Flags().BoolVar(&urlSafe, "url", false, "Use URL-safe encoding")

	decodeCmd := &cobra.Command{
		Use:   "decode [text]",
		Short: "Decode from Base64",
		RunE: func(cmd *cobra.Command, args []string) error {
			in, err := readInput(args)
			if err != nil {
				return err
			}
			var out []byte
			if urlSafe {
				res, err := cryptzb64.UrlDecode(string(in))
				if err != nil {
					return err
				}
				out = res.Bytes()
			} else {
				res, err := cryptzb64.StdDecode(string(in))
				if err != nil {
					return err
				}
				out = res.Bytes()
			}
			fmt.Fprint(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
	decodeCmd.Flags().BoolVar(&urlSafe, "url", false, "Use URL-safe encoding")

	cmd.AddCommand(encodeCmd, decodeCmd)
	return cmd
}
