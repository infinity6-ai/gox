package bqzjob_test

import (
	"testing"

	"github.com/infinity6-ai/gox/bqz/bqztable"
	"github.com/infinity6-ai/gox/commonz/validation/checked/tuchecked"
)

func TestUnitChecked(t *testing.T) {
	tuchecked.Check(tuchecked.Table[*bqztable.Table, string]{
		Create: func(v string) *bqztable.Table {
			ret := bqztable.New(v)
			return &ret
		},
		Valids:   []string{"a1", "b_1", "c1"},
		Invalids: []string{"", "a", "a-b", "ab-"},
	})
}
