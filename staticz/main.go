package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
			out := internal.Check2(cmd.Flags().GetString("out"))
			if out == "-" {
				opts.Out = os.Stdout
			} else {
				out = internal.Check2(filepath.Abs(out))
				os.MkdirAll(filepath.Dir(out), os.ModePerm)
				opts.Out = internal.Check2(os.Create(out))
			}
			staticz.Generate(ctx, opts)
		},
	}
	cmd.PersistentFlags().String("imp", "github.com/infinity6-ai/gox/commonz/staticz", "import module")
	cmd.PersistentFlags().String("code", "staticz.Instance().SetCode", "code")
	cmd.PersistentFlags().String("dir", "stzfiles", "dir")
	cmd.PersistentFlags().String("out", "internal/stzfiles/stzfiles_gen.go", "outfile")

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
