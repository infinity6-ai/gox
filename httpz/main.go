package main

import (
	"context"

	"github.com/infinity6-ai/gox/httpz/cmdhttpz"
)

func main() {
	ctx := context.Background()
	rootCmd := cmdhttpz.Prepare(ctx)
	cmdhttpz.Execute(rootCmd)
}
