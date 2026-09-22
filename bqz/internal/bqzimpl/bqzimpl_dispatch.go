package bqzimpl

import (
	"context"
	"fmt"

	"github.com/infinity6-ai/gox/bqz/bqz"
	"github.com/infinity6-ai/gox/bqz/bqzjob"
)

// Dispatch implements [bqz.Service].
func (b *BqzServiceImpl) Dispatch(ctx context.Context, query *bqz.Query) (*bqz.Job, error) {
	q := b.c.Query(query.Query)
	job, err := q.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run query job: %w", err)
	}
	return &bqz.Job{
		Id: bqzjob.New(job.ID()),
	}, nil
}
