package main

import (
	"context"

	"github.com/infinity6-ai/gox/cryptz/cmdcryptz"
)

func main() {
	ctx := context.Background()
	rootCmd := cmdcryptz.Prepare(ctx)
	cmdcryptz.Execute(rootCmd)
}
