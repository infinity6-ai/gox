package cmdcryptz

import (
	"fmt"

	"github.com/infinity6-ai/gox/cryptz/cryptzb32"
	"github.com/spf13/cobra"
)

func newB32Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "b32",
		Short: "Base32 encode/decode",
	}

	encodeCmd := &cobra.Command{
		Use:   "encode [text]",
		Short: "Encode to Base32",
		RunE: func(cmd *cobra.Command, args []string) error {
			in, err := readInput(args)
			if err != nil {
				return err
			}
			out := cryptzb32.Encode(in)
			fmt.Print(string(out.Bytes()))
			return nil
		},
	}

	decodeCmd := &cobra.Command{
		Use:   "decode [text]",
		Short: "Decode from Base32",
		RunE: func(cmd *cobra.Command, args []string) error {
			in, err := readInput(args)
			if err != nil {
				return err
			}
			out, err := cryptzb32.Decode(string(in))
			if err != nil {
				return err
			}
			fmt.Print(string(out.Bytes()))
			return nil
		},
	}

	cmd.AddCommand(encodeCmd, decodeCmd)
	return cmd
}
