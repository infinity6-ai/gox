package main

import (
	"context"

	"github.com/infinity6-ai/gox/commonz/trokerz/cmdtrokerz"
)

func main() {
	ctx := context.Background()
	cmd := cmdtrokerz.Prepare(ctx, nil)
	cmdtrokerz.Execute(ctx, cmd)
}
