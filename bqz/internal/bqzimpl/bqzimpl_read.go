package bqzimpl

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"

	"github.com/infinity6-ai/gox/bqz/bqz"
	"github.com/infinity6-ai/gox/bqz/bqzjob"
)

// Read implements [bqz.Service].
func (b *BqzServiceImpl) Read(ctx context.Context, job bqzjob.Job) (bqz.Iterator, error) {
	j, err := b.c.JobFromID(ctx, job.Get())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve job %s: %w", job.Get(), err)
	}
	it, err := j.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read results for job %s: %w", job.Get(), err)
	}

	return func(ctx context.Context, v any) (bool, error) {
		// Read raw bigquery values instead of letting the default bigquery library load it,
		// so that we can support custom Null types like dataz.NullString.
		var values []bigquery.Value
		err := it.Next(&values)
		if err == iterator.Done {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("failed to read next row: %w", err)
		}

		// map []bigquery.Value to v using it.Schema
		if err := LoadRow(v, it.Schema, values); err != nil {
			return false, fmt.Errorf("failed to map bigquery row: %w", err)
		}

		return true, nil
	}, nil
}
