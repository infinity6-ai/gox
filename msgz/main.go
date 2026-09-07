package main

import (
	"context"

	"github.com/infinity6-ai/gox/msgz/cmdmsgz"
)

func main() {
	ctx := context.Background()
	rootCmd := cmdmsgz.Prepare(ctx)
	cmdmsgz.Execute(rootCmd)
}
