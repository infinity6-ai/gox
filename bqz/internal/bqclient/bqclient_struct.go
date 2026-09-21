package bqclient

import (
	"context"

	"cloud.google.com/go/bigquery"
)

type ClientOptions struct {
	Project string
}

type Client struct {
	client *bigquery.Client
}

func New(ctx context.Context, opts ClientOptions) *Client {
	panic("implement it")
}
