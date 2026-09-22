package bqz

import "time"

type Labels map[string]string

type Dataset struct {
	Name                       string
	Labels                     map[string]string
	DefaultPartitionExpiration time.Duration
}

type ExternalTable struct {
	Dataset   string
	Table     string
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
