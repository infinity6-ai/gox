package staticz

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
	return rootCmd
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
