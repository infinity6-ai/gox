package bqclient

import (
	"context"
	"fmt"
)

func (c *Client) WaitFor(ctx context.Context, job JobId) error {
	j, err := c.client.JobFromID(ctx, string(job))
	if err != nil {
		return fmt.Errorf("failed to retrieve job %s: %w", job, err)
	}
	status, err := j.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for job %s: %w", job, err)
	}
	if err := status.Err(); err != nil {
		return fmt.Errorf("job %s failed: %w", job, err)
	}
	return nil
}
