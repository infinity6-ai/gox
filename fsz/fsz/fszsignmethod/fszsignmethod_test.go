package fszsignmethod_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/validation/checked/tuchecked"
	"github.com/infinity6-ai/gox/fsz/fsz/fszsignmethod"
)

func TestUnitChecked(t *testing.T) {
	tuchecked.Check(tuchecked.Table[*fszsignmethod.SignMethod, string]{
		Create: func(v string) *fszsignmethod.SignMethod {
			return new(fszsignmethod.MustParse(v))
		},
		Valids: []string{
			fszsignmethod.SignMethodGet.Get(),
			fszsignmethod.SignMethodPut.Get(),
			fszsignmethod.SignMethodDelete.Get(),
		},
		Invalids: []string{"", "a_a", "_1a", "a1_", "patch", "PATCH", "options", "OPTIONS"},
	})
}
