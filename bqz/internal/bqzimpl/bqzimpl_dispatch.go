package bqzimpl

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/infinity6-ai/gox/bqz/bqz"
	"github.com/infinity6-ai/gox/bqz/bqzjob"
)

// Dispatch implements [bqz.Service].
func (b *BqzServiceImpl) Dispatch(ctx context.Context, query *bqz.Query) (*bqz.Job, error) {
	q := b.c.Query(query.Query)
	if len(query.Binds) > 0 {
		q.Parameters = make([]bigquery.QueryParameter, 0, len(query.Binds))
		for k, v := range query.Binds {
			q.Parameters = append(q.Parameters, bigquery.QueryParameter{
				Name:  k,
				Value: v,
			})
		}
	}
	job, err := q.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run query job: %w", err)
	}
	return &bqz.Job{
		Id: bqzjob.New(job.ID()),
	}, nil
}
