package bqzjob_test

import (
	"testing"

	"github.com/infinity6-ai/gox/bqz/bqzjob"
	"github.com/infinity6-ai/gox/commonz/validation/checked/tuchecked"
)

func TestUnitChecked(t *testing.T) {
	tuchecked.Check(tuchecked.Table[*bqzjob.Job, string]{
		Create: func(v string) *bqzjob.Job {
			ret := bqzjob.New(v)
			return &ret
		},
		Valids:   []string{"a1", "b_1", "c1", "a", "a-b", "ab-", "A_b-1"},
		Invalids: []string{"", "a.b", "a/b", "a b"},
	})
}
