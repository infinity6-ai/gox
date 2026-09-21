package bqclient

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
)

type DatasetOptions struct {
	Dataset  string
	Metadata *bigquery.DatasetMetadata
}

func (c *Client) CreateDataset(ctx context.Context, opts DatasetOptions) error {
	ds := c.client.Dataset(opts.Dataset)
	_, err := ds.Metadata(ctx)
	if err == nil {
		return nil
	}
	if ParseErrorCode(err) != 404 {
		return fmt.Errorf("error loading dataset: %#v, %w", opts, err)
	}
	err = ds.Create(ctx, opts.Metadata)
	if err == nil {
		return nil
	}
	if ParseErrorCode(err) != 409 {
		return fmt.Errorf("error creating dataset: %#v, %w", opts, err)
	}
	return nil
}
