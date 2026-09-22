package bqzimpl

import (
	"context"
	"fmt"

	"github.com/infinity6-ai/gox/bqz/bqzjob"
)

// WaitFor implements [bqz.Service].
func (b *BqzServiceImpl) WaitFor(ctx context.Context, job bqzjob.Job) error {
	j, err := b.c.JobFromID(ctx, job.Get())
	if err != nil {
		return fmt.Errorf("failed to retrieve job %s: %w", job.Get(), err)
	}
	status, err := j.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for job %s: %w", job.Get(), err)
	}
	if err := status.Err(); err != nil {
		return fmt.Errorf("job %s failed: %w", job.Get(), err)
	}
	return nil
}
