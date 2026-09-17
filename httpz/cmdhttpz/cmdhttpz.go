// Package cmdhttpz provides command-line interface definitions and commands for httpz.
package cmdhttpz

import (
	"context"
	"fmt"
	"os"

	"github.com/infinity6-ai/gox/versionz/version"
	"github.com/spf13/cobra"
)

// Prepare initializes and returns the root cobra command for httpz.
func Prepare(ctx context.Context) *cobra.Command {
	var rootCmd = &cobra.Command{
		Version: version.Version(),
		Use:     "httpz",
		Short:   "httpz",
	}

	rootCmd.AddCommand(&cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Version())
		},
	})

	return rootCmd
}

// Execute executes the given root cobra command and terminates the process on error.
func Execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
