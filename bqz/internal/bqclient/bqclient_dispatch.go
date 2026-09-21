package bqclient

import (
	"context"
	"fmt"
)

type QueryOptions struct {
	Query string
}

func (c *Client) Dispatch(ctx context.Context, query QueryOptions) (JobId, error) {
	q := c.client.Query(query.Query)
	job, err := q.Run(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to run query job: %w", err)
	}
	return JobId(job.ID()), nil
}
