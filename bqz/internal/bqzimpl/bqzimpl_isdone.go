package bqzimpl

import (
	"context"
	"fmt"

	"github.com/infinity6-ai/gox/bqz/bqzjob"
)

// IsDone implements [bqz.Service].
func (b *BqzServiceImpl) IsDone(ctx context.Context, job bqzjob.Job) (bool, error) {
	j, err := b.c.JobFromID(ctx, job.Get())
	if err != nil {
		return false, fmt.Errorf("failed to retrieve job %s: %w", job.Get(), err)
	}
	status, err := j.Status(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get status for job %s: %w", job.Get(), err)
	}
	return status.Done(), nil
}
