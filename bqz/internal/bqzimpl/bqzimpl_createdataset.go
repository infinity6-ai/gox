package bqzimpl

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/infinity6-ai/gox/bqz/bqz"
)

// CreateDataset implements [bqz.Service].
func (b *BqzServiceImpl) CreateDataset(ctx context.Context, dataset *bqz.Dataset) error {
	ds := b.c.Dataset(dataset.Name)
	_, err := ds.Metadata(ctx)
	if err == nil {
		return nil
	}
	if ParseErrorCode(err) != 404 {
		return fmt.Errorf("error loading dataset metadata for %s: %w", dataset.Name, err)
	}

	meta := &bigquery.DatasetMetadata{
		Labels:                 dataset.Labels,
		DefaultTableExpiration: dataset.DefaultTableExpiration,
	}

	err = ds.Create(ctx, meta)
	if err == nil {
		return nil
	}
	if ParseErrorCode(err) != 409 {
		return fmt.Errorf("error creating dataset %s: %w", dataset.Name, err)
	}
	return nil
}
