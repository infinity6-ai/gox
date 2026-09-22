package bqzimpl

import (
	"context"

	"github.com/infinity6-ai/gox/bqz/bqz"
)

// WaitFor implements [bqz.Service].
func (b *BqzServiceImpl) WaitFor(ctx context.Context, job *bqz.Job) error {
	panic("unimplemented")
}
