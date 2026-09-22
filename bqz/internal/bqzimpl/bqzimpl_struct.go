package bqzimpl

import (
	"context"
	"errors"

	"cloud.google.com/go/bigquery"
	"github.com/infinity6-ai/gox/bqz/bqz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"google.golang.org/api/googleapi"
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

func ParseErrorCode(err error) int {
	if err == nil {
		return -1
	}
	var e *googleapi.Error
	ok := errors.As(err, &e)
	if !ok {
		return -1
	}
	return e.Code
}
