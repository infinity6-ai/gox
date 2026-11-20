// Package cmdbrwjsz provides command-line interface definitions and commands for brwjsz.
package cmdbrwjsz

import (
	"context"
	"fmt"
	"os"

	"github.com/infinity6-ai/gox/brwjsz/internal/cmd/cmdexec"
	"github.com/infinity6-ai/gox/brwjsz/version"
	"github.com/spf13/cobra"
)

// Prepare initializes and returns the root cobra command for brwjsz.
func Prepare(ctx context.Context) *cobra.Command {
	var rootCmd = &cobra.Command{
		Version: version.Version(),
		Use:     "brwjsz",
		Short:   "brwjsz browser execution and IPC bridge",
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the current version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Version())
		},
	})

	cmdexec.Prepare(ctx, rootCmd)

	return rootCmd
}

// Execute executes the given root cobra command with the provided context and terminates the process on error.
func Execute(ctx context.Context, rootCmd *cobra.Command) {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
