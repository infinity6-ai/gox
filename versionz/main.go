package main

import (
	"context"

	"github.com/infinity6-ai/gox/versionz/cmdversionz"
)

func main() {
	ctx := context.Background()
	rootCmd := cmdversionz.Prepare(ctx)
	cmdversionz.Execute(ctx, rootCmd)
}
