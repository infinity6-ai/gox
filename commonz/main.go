//go:generate go run github.com/infinity6-ai/gox/commonz@latest staticz generate --out internal/stzfiles/stzfiles.blob.go
package main

import (
	"context"

	"github.com/infinity6-ai/gox/commonz/cmdcommonz"
)

func main() {
	ctx := context.Background()
	rootCmd := cmdcommonz.Prepare(ctx)
	cmdcommonz.Execute(ctx, rootCmd)
}
