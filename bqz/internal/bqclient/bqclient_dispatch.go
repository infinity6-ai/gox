package bqclient

import (
	"context"
)

type QueryOptions struct {
	Query string
}

func (c *Client) Dispatch(ctx context.Context, query QueryOptions) (JobId, error) {
	panic("implement")
}
