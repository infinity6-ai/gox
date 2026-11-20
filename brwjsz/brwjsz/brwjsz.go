package brwjsz

import (
	"context"

	"github.com/infinity6-ai/gox/commonz/execz"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
)

type Options struct {
	Command execz.Options
	Server  httpzserver.Options
}

type Browser struct {
}

func New(ctx context.Context, opts Options) *Browser {
	panic("implement it")
}
