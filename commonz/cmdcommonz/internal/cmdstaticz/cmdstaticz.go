package cmdstaticz

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzwriter"
	"github.com/infinity6-ai/gox/versionz/version"
	"github.com/spf13/cobra"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

func Prepare(ctx context.Context, parent *cobra.Command) *cobra.Command {
	var cmd = &cobra.Command{
		Version: version.Version(),
		Use:     "staticz",
		Short:   "staticz",
	}
	parent.AddCommand(cmd)

	prepareGenerateCmd(cmd)

	return cmd
}

func prepareGenerateCmd(parent *cobra.Command) {
	var opts staticzwriter.GenerateOptions
	var out string

	cmd := &cobra.Command{
		Use: "generate",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			var buf bytes.Buffer
			opts.Out = &buf
			staticzwriter.Generate(ctx, opts)

			var w io.Writer = cmd.OutOrStdout()
			if out != "-" {
				out = errorz.Check2(filepath.Abs(out))
				logger.Info(ctx, "stfiles generation", map[string]any{"out": out, "size": buf.Len()})
				errorz.Check(os.MkdirAll(filepath.Dir(out), 0o755))
				f := errorz.Check2(os.Create(out))
				defer f.Close()
				w = f
			}
			_, err := buf.WriteTo(w)
			errorz.Check(err)
		},
	}
	cmd.Flags().StringVar(&opts.Imp, "imp", "", "import module")
	cmd.Flags().StringVar(&opts.Code, "code", "", "code")
	cmd.Flags().StringVar(&opts.Dir, "dir", "", "dir")
	cmd.Flags().StringVar(&out, "out", "internal/stzfiles/stzfiles.gen.go", "outfile")

	parent.AddCommand(cmd)
}
