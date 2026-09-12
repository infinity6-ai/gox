package bucketz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/validation/checked/tuchecked"
	"github.com/infinity6-ai/gox/fsz/bucketz"
)

func TestUnitChecked(t *testing.T) {
	tuchecked.Check(tuchecked.Table[*bucketz.Bucket, string]{
		Create: func(v string) *bucketz.Bucket {
			ret := bucketz.Id(v)
			return &ret
		},
		Valids:   []string{"a1a", "b2b", "c-c"},
		Invalids: []string{"", "a_a", "_1a", "a1_"},
	})
}
