package bqz

import (
	"context"
	"time"

	"github.com/infinity6-ai/gox/bqz/bqzdataset"
	"github.com/infinity6-ai/gox/bqz/bqzjob"
	"github.com/infinity6-ai/gox/bqz/bqztable"
)

type Labels map[string]string

type Dataset struct {
	Name                   string
	Labels                 map[string]string
	DefaultTableExpiration time.Duration
}

type ExternalTable struct {
	Dataset   bqzdataset.Dataset
	Table     bqztable.Table
	Uri       string
	HiveParts []string
	Schema    any
	Labels    map[string]string
}

type Query struct {
	Query string
}

type Job struct {
	Id bqzjob.Job
}

type Iterator func(ctx context.Context, v any) (bool, error)

type Service interface {
	CreateDataset(ctx context.Context, dataset *Dataset) error
	CreateExternalTable(ctx context.Context, table *ExternalTable) error
	Dispatch(ctx context.Context, query *Query) error
	WaitFor(ctx context.Context, job *Job) error
	IsDone(ctx context.Context, job *Job) error
	Read(ctx context.Context, job *Job) (Iterator, error)
}
