package main

import (
	"context"

	"github.com/infinity6-ai/gox/routez/cmdroutez"
)

func main() {
	ctx := context.Background()
	rootCmd := cmdroutez.Prepare(ctx)
	cmdroutez.Execute(rootCmd)
}
