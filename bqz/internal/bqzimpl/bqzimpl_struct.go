package bqzimpl

import (
	"context"

	"cloud.google.com/go/bigquery"
	"github.com/infinity6-ai/gox/bqz/bqz"
	"github.com/infinity6-ai/gox/commonz/errorz"
)

type BqzServiceImpl struct {
	c *bigquery.Client
}

func New(ctx context.Context, opts bqz.ClientOptions) *BqzServiceImpl {
	client, err := bigquery.NewClient(ctx, opts.Project)
	errorz.Check(err)
	return &BqzServiceImpl{
		c: client,
	}
}
