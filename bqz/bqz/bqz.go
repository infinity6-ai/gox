package bqz

import (
	"time"

	"github.com/infinity6-ai/gox/bqz/bqzdataset"
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
	Id string
}
