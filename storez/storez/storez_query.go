package storez

import "github.com/infinity6-ai/gox/commonz/gobz"

type Filter struct {
	Field string
	Op    string
	Value any
}

type OrderBy struct {
	Field string
	Desc  bool
}

type Query struct {
	Table string

	StartCursor string
	Limit       int

	Filters []Filter

	OrderBys []OrderBy

	Projections []string
}

func (q *Query) AddFilterPrefix(name string, prefix string) {
	q.Filters = append(q.Filters, Filter{Field: name, Op: ">=", Value: prefix}, Filter{Field: name, Op: "<", Value: prefix + "\ufffd"})
}

func (q *Query) AddFilterString(name string, op string, value string) {
	if value != "" {
		q.Filters = append(q.Filters, Filter{Field: name, Op: op, Value: value})
	}
}

func (q *Query) Clone() *Query {
	return gobz.MustClone(q, &Query{})
}
