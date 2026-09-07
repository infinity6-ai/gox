package cmdcommonz

import (
	"context"
	"fmt"
	"os"

	"github.com/infinity6-ai/gox/commonz/cmdcommonz/internal/cmdstaticz"
	"github.com/spf13/cobra"
)

func Prepare(ctx context.Context) *cobra.Command {
	var rootCmd = &cobra.Command{
		Version: "dev",
		Use:     "commonz",
		Short:   "commonz",
	}

	// rootCmd.AddCommand(&cobra.Command{
	// 	Use: "version",
	// 	Run: func(cmd *cobra.Command, args []string) {
	// 		// fmt.Printf("%s\n", version.Version())
	// 	},
	// })

	cmdstaticz.Prepare(ctx, rootCmd)

	return rootCmd
}

func Execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
