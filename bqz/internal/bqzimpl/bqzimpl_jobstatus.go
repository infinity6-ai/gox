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
		return bqz.JobStatusUnspecified, fmt.Errorf("failed to retrieve job %s: %w", job.Get(), err)
	}
	status, err := j.Status(ctx)
	if status == nil {
		if err != nil {
			return bqz.JobStatusUnspecified, fmt.Errorf("failed to get status for job %s: %w", job.Get(), err)
		}
		return bqz.JobStatusUnspecified, nil
	}

	var jobStatus bqz.JobStatus
	switch status.State {
	case bigquery.Pending:
		jobStatus = bqz.JobStatusCreated
	case bigquery.Running:
		jobStatus = bqz.JobStatusRunning
	case bigquery.Done:
		jobStatus = bqz.JobStatusDone
	default:
		jobStatus = bqz.JobStatusNotFound
	}

	if err != nil {
		return jobStatus, fmt.Errorf("failed to get status for job %s: %w", job.Get(), err)
	}

	if err := status.Err(); err != nil {
		return jobStatus, fmt.Errorf("job failed: %w", err)
	}

	return jobStatus, nil
}
