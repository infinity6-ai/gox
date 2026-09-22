package bqzdatasetid_test

import (
	"testing"

	"github.com/infinity6-ai/gox/bqz/bqz/bqzdatasetid"
	"github.com/infinity6-ai/gox/commonz/validation/checked/tuchecked"
)

func TestUnitChecked(t *testing.T) {
	tuchecked.Check(tuchecked.Table[*bqzdatasetid.DatasetId, string]{
		Create: func(v string) *bqzdatasetid.DatasetId {
			ret := bqzdatasetid.New(v)
			return &ret
		},
		Valids:   []string{"a1", "b_1", "c1"},
		Invalids: []string{"", "a", "a-b", "ab-"},
	})
}
