package main

import (
	"context"

	"github.com/infinity6-ai/gox/brwjsz/cmdbrwjsz"
)

func main() {
	ctx := context.Background()
	rootCmd := cmdbrwjsz.Prepare(ctx)
	cmdbrwjsz.Execute(ctx, rootCmd)
}
