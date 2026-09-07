package main

import (
	"context"

	"github.com/infinity6-ai/gox/fsz/cmdfsz"
)

func main() {
	ctx := context.Background()
	rootCmd := cmdfsz.Prepare(ctx)
	cmdfsz.Execute(rootCmd)
}
