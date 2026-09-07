package main

import (
	"context"

	"github.com/infinity6-ai/gox/schemaz/cmdschemaz"
)

func main() {
	ctx := context.Background()
	rootCmd := cmdschemaz.Prepare(ctx)
	cmdschemaz.Execute(rootCmd)
}
