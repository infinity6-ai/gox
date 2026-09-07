package main

import (
	"context"
	"fmt"
	"os"

	"github.com/infinity6-ai/gox/staticz/internal"
	"github.com/infinity6-ai/gox/staticz/staticz"
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
			opts := staticz.GenerateOptions{}
			opts.Imp = internal.Check2(cmd.Flags().GetString("imp"))
			opts.Code = internal.Check2(cmd.Flags().GetString("code"))
			opts.Dir = internal.Check2(cmd.Flags().GetString("dir"))
			staticz.Generate(ctx, opts)
		},
	}
	cmd.PersistentFlags().String("imp", "import module", "")
	cmd.PersistentFlags().String("code", "code", "")
	cmd.PersistentFlags().String("dir", "dir", "")
	cmd.MarkPersistentFlagRequired("imp")
	cmd.MarkPersistentFlagRequired("code")
	cmd.MarkPersistentFlagRequired("dir")

	parent.AddCommand(cmd)
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
