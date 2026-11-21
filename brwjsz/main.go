//go:generate go run github.com/infinity6-ai/gox/commonz@latest staticz generate --out internal/stzfiles/stzfiles.blob.go
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
