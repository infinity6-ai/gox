package bqzservice

import (
	"context"

	"github.com/infinity6-ai/gox/bqz/bqz"
	"github.com/infinity6-ai/gox/bqz/internal/bqzimpl"
)

func New(ctx context.Context) bqz.Service {
	return bqzimpl.New()
}
