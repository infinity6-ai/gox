//go:generate go run github.com/infinity6-ai/gox/commonz generate
package main

import (
	"context"

	"github.com/infinity6-ai/gox/commonz/cmdcommonz"
)

func main() {
	ctx := context.Background()
	rootCmd := cmdcommonz.Prepare(ctx)
	cmdcommonz.Execute(rootCmd)
}
