package bqzimpl

import (
	"context"
	"fmt"

	"google.golang.org/api/iterator"

	"github.com/infinity6-ai/gox/bqz/bqz"
	"github.com/infinity6-ai/gox/bqz/bqzjob"
)

// Read implements [bqz.Service].
func (b *BqzServiceImpl) Read(ctx context.Context, job bqzjob.Job) (bqz.Iterator, error) {
	j, err := b.c.JobFromID(ctx, job.Get())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve job %s: %w", job.Get(), err)
	}
	it, err := j.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read results for job %s: %w", job.Get(), err)
	}

	return func(ctx context.Context, v any) (bool, error) {
		err := it.Next(v)
		if err == iterator.Done {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("failed to read next row: %w", err)
		}
		return true, nil
	}, nil
}
