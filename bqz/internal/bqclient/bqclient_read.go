package bqclient

import (
	"context"

	"cloud.google.com/go/bigquery"
)

func (c *Client) Read(ctx context.Context, job JobId) (*bigquery.RowIterator, error) {
	panic("implement it")
}
