package bqzimpl

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/infinity6-ai/gox/bqz/bqz"
)

// Dispatch implements [bqz.Service].
func (b *BqzServiceImpl) Dispatch(ctx context.Context, query *bqz.Query) error {
	q := b.c.Query(query.Query)
	q.JobID = query.Job.Get()
	if query.RunningTimeout > 0 {
		q.JobTimeout = query.RunningTimeout
	}
	if len(query.Binds) > 0 {
		q.Parameters = make([]bigquery.QueryParameter, 0, len(query.Binds))
		for k, v := range query.Binds {
			q.Parameters = append(q.Parameters, bigquery.QueryParameter{
				Name:  k,
				Value: v,
			})
		}
	}
	_, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to run query job: %w", err)
	}
	return nil
}
