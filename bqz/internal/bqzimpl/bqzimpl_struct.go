package bqzimpl

import (
	"context"

	"github.com/infinity6-ai/gox/bqz/bqz"
)

type BqzServiceImpl struct {
}

// CreateExternalTable implements [bqz.Service].
func (b *BqzServiceImpl) CreateExternalTable(ctx context.Context, table *bqz.ExternalTable) error {
	panic("unimplemented")
}

// Dispatch implements [bqz.Service].
func (b *BqzServiceImpl) Dispatch(ctx context.Context, query *bqz.Query) error {
	panic("unimplemented")
}

// IsDone implements [bqz.Service].
func (b *BqzServiceImpl) IsDone(ctx context.Context, job *bqz.Job) error {
	panic("unimplemented")
}

// Read implements [bqz.Service].
func (b *BqzServiceImpl) Read(ctx context.Context, job *bqz.Job) (bqz.Iterator, error) {
	panic("unimplemented")
}

// WaitFor implements [bqz.Service].
func (b *BqzServiceImpl) WaitFor(ctx context.Context, job *bqz.Job) error {
	panic("unimplemented")
}

func New() *BqzServiceImpl {
	return &BqzServiceImpl{}
}
