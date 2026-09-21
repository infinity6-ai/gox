package main

import (
	"context"

	"github.com/infinity6-ai/gox/bqz/cmdbqz"
)

func main() {
	ctx := context.Background()
	rootCmd := cmdbqz.Prepare(ctx)
	cmdbqz.Execute(rootCmd)
}
