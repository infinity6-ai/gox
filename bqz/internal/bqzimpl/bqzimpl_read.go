package bqzimpl

import (
	"context"

	"github.com/infinity6-ai/gox/bqz/bqz"
)

// Read implements [bqz.Service].
func (b *BqzServiceImpl) Read(ctx context.Context, job *bqz.Job) (bqz.Iterator, error) {
	panic("unimplemented")
}
