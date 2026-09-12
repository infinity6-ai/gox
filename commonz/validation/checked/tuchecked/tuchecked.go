package tuchecked

import (
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"go.code.infinity6.ai/platform/errorz"
)

type Table[T comparable, V comparable] struct {
	Create func(v V) T
	Valids []V
	// Invalids []V
}

func Check[T comparable, V comparable](table Table[T, V]) {
	for idx, valid := range table.Valids {
		var t T
		err := errorz.UnpanicV(func() {
			t = table.Create(valid)
		})
		checker.Nil(err, "panic while creating [idx=%d]: %v", idx, valid)
		checker.True(t == table.Create(valid), "creating with same value must be equal, expected: %v, but was: %v", valid, valid)
		// for _, eq := range eqs {
		// checker.True(valid == eq, "it is not equal, expected: %v, but was: %v", valid, eq)
		// checker.True(v == table.Create(eq), "it is not equal, expected: %v, but was: %v", valid, eq)
		// }
	}
}

// func Check[V comparable](create func(v V) Checker[V], valids []V, invalids []V) {
// 	for i, valid := range valids {
// 		v := create(valid)
// 		// check it is a type of Value
// 	}
// }
