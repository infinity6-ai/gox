package bqzimpl

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/infinity6-ai/gox/bqz/bqz"
	"github.com/infinity6-ai/gox/bqz/bqzjob"
)

// JobStatus implements [bqz.Service].
func (b *BqzServiceImpl) JobStatus(ctx context.Context, job bqzjob.Job) (bqz.JobStatus, error) {
	j, err := b.c.JobFromID(ctx, job.Get())
	if err != nil {
		if ParseErrorCode(err) == 404 {
			return bqz.JobStatusNotFound, nil
		}
		return bqz.JobStatusNotFound, fmt.Errorf("failed to retrieve job %s: %w", job.Get(), err)
	}
	status, err := j.Status(ctx)
	if err != nil {
		return bqz.JobStatusNotFound, fmt.Errorf("failed to get status for job %s: %w", job.Get(), err)
	}

	switch status.State {
	case bigquery.Pending:
		return bqz.JobStatusCreated, nil
	case bigquery.Running:
		return bqz.JobStatusRunning, nil
	case bigquery.Done:
		return bqz.JobStatusDone, nil
	default:
		return bqz.JobStatusNotFound, nil
	}
}
