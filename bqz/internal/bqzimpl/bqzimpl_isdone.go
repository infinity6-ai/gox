package bqzimpl

import (
	"context"
	"fmt"

	"github.com/infinity6-ai/gox/bqz/bqz"
)

// IsDone implements [bqz.Service].
func (b *BqzServiceImpl) IsDone(ctx context.Context, job *bqz.Job) (bool, error) {
	j, err := b.c.JobFromID(ctx, job.Id.Get())
	if err != nil {
		return false, fmt.Errorf("failed to retrieve job %s: %w", job.Id.Get(), err)
	}
	status, err := j.Status(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get status for job %s: %w", job.Id.Get(), err)
	}
	return status.Done(), nil
}
