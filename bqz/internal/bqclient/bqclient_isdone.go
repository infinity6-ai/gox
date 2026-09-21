package bqclient

import (
	"context"
	"fmt"
)

func (c *Client) IsDone(ctx context.Context, job JobId) (bool, error) {
	j, err := c.client.JobFromID(ctx, string(job))
	if err != nil {
		return false, fmt.Errorf("failed to retrieve job %s: %w", job, err)
	}
	status, err := j.Status(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get status for job %s: %w", job, err)
	}
	return status.Done(), nil
}
