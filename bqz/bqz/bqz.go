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
	Job   bqzjob.Job
	Query string
	Binds map[string]any
}

type Iterator func(ctx context.Context, v any) (bool, error)

type ClientOptions struct {
	Project string
}

type JobStatus string

const (
	JobStatusNotFound JobStatus = "NOTFOUND"
	JobStatusCreated  JobStatus = "CREATED"
	JobStatusRunning  JobStatus = "RUNNING"
	JobStatusDone     JobStatus = "DONE"
)

type Service interface {
	CreateDataset(ctx context.Context, dataset *Dataset) error
	CreateExternalTable(ctx context.Context, table *ExternalTable) error
	Dispatch(ctx context.Context, query *Query) error
	WaitFor(ctx context.Context, job bqzjob.Job) error
	JobStatus(ctx context.Context, job bqzjob.Job) (JobStatus, error)
	Read(ctx context.Context, job bqzjob.Job) (Iterator, error)
	TableExists(ctx context.Context, dataset bqzdataset.Dataset, table bqztable.Table) (bool, error)
	DropTable(ctx context.Context, dataset bqzdataset.Dataset, table bqztable.Table) error
}
