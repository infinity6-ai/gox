package bqclientdataset

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/infinity6-ai/gox/bqz/internal/bqclient/bqclienterror"
)

type Options struct {
	Dataset  string
	Metadata *bigquery.DatasetMetadata
}

func CreateDataset(ctx context.Context, c *bigquery.Client, opts Options) error {
	ds := c.Dataset(opts.Dataset)
	_, err := ds.Metadata(ctx)
	if err == nil {
		return nil
	}
	if bqclienterror.ParseErrorCode(err) != 404 {
		return fmt.Errorf("error loading dataset: %#v, %w", opts, err)
	}
	err = ds.Create(ctx, opts.Metadata)
	if err == nil {
		return nil
	}
	if bqclienterror.ParseErrorCode(err) != 409 {
		return fmt.Errorf("error creating dataset: %#v, %w", opts, err)
	}
	return nil
}
