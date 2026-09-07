package cmdstaticz

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/staticz/staticz"
	"github.com/spf13/cobra"
)

func Prepare(ctx context.Context, parent *cobra.Command) *cobra.Command {
	var cmd = &cobra.Command{
		Version: "dev",
		Use:     "staticz",
		Short:   "staticz",
	}
	parent.AddCommand(cmd)

	prepareGenerateCmd(ctx, cmd)

	return cmd
}

func prepareGenerateCmd(ctx context.Context, parent *cobra.Command) {
	cmd := &cobra.Command{
		Use: "generate",
		Run: func(cmd *cobra.Command, args []string) {
			opts := staticz.GenerateOptions{}
			opts.Imp = errorz.Check2(cmd.Flags().GetString("imp"))
			opts.Code = errorz.Check2(cmd.Flags().GetString("code"))
			opts.Dir = errorz.Check2(cmd.Flags().GetString("dir"))
			out := errorz.Check2(cmd.Flags().GetString("out"))
			if out == "-" {
				opts.Out = os.Stdout
			} else {
				out = errorz.Check2(filepath.Abs(out))
				os.MkdirAll(filepath.Dir(out), os.ModePerm)
				opts.Out = errorz.Check2(os.Create(out))
			}
			staticz.Generate(ctx, opts)
		},
	}
	cmd.PersistentFlags().String("imp", "github.com/infinity6-ai/gox/commonz/staticz", "import module")
	cmd.PersistentFlags().String("code", "staticz.Instance().SetCode", "code")
	cmd.PersistentFlags().String("dir", "stzfiles", "dir")
	cmd.PersistentFlags().String("out", "internal/stzfiles/stzfiles.gen.go", "outfile")

	parent.AddCommand(cmd)
}

func Execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
