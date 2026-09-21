package bqz

type ExternalTable struct {
	Dataset   string
	Table     string
	Uri       string
	HiveParts []string
	Schema    any
}
