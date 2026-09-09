package main

import (
	"context"

	"github.com/infinity6-ai/gox/storez/cmdstorez"
)

func main() {
	ctx := context.Background()
	rootCmd := cmdstorez.Prepare(ctx)
	cmdstorez.Execute(rootCmd)
}
