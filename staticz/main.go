package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const Version = "v0.0.1"

func Prepare(ctx context.Context) *cobra.Command {
	var rootCmd = &cobra.Command{
		Version: Version,
		Use:     "staticz",
		Short:   "staticz",
	}

	prepareGenerateCmd(ctx, rootCmd)

	return rootCmd
}

func prepareGenerateCmd(ctx context.Context, parent *cobra.Command) {
	cmd := &cobra.Command{
		Use: "generate",
		Run: func(cmd *cobra.Command, args []string) {
			opts := GenerateOptions{}
			opts.Imp = check2(cmd.Flags().GetString("imp"))
			opts.Code = check2(cmd.Flags().GetString("code"))
			Generate(ctx, opts)
		},
	}
	cmd.PersistentFlags().String("imp", "import module", "")
	cmd.PersistentFlags().String("code", "code", "")
	cmd.MarkPersistentFlagRequired("imp")
	cmd.MarkPersistentFlagRequired("code")

	parent.AddCommand(cmd)
}

func check2[T any](s T, err error) T {
	if err != nil {
		panic(err)
	}
	return s
}

func Execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	ctx := context.Background()
	rootCmd := Prepare(ctx)
	Execute(rootCmd)
}
