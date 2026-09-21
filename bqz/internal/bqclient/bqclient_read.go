package bqclient

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
)

func (c *Client) Read(ctx context.Context, job JobId) (*bigquery.RowIterator, error) {
	j, err := c.client.JobFromID(ctx, string(job))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve job %s: %w", job, err)
	}
	it, err := j.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read results for job %s: %w", job, err)
	}
	return it, nil
}
