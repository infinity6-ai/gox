package tuchecked

import (
	"github.com/infinity6-ai/gox/commonz/validation/checked"
)

type Table[T checked.Checker[V], V comparable] struct {
	Create func(v V) T
	// Valids map[V][]V
	// Invalids []V
}

func Check[T checked.Checker[V], V comparable](table Table[T, V]) {
	// for valid, eqs := range table.Valids {
	// 	var v checked.Checker[V]
	// 	err := errorz.UnpanicV(func() {
	// 		v = table.Create(valid)
	// 	})
	// 	checker.Nil(err, "panic while creating: %v", valid)
	// 	checker.True(v == table.Create(valid), "creating with same value must be equal, expected: %v, but was: %v", valid, valid)
	// 	for _, eq := range eqs {
	// 		checker.True(valid == eq, "it is not equal, expected: %v, but was: %v", valid, eq)
	// 		checker.True(v == table.Create(eq), "it is not equal, expected: %v, but was: %v", valid, eq)
	// 	}
	// }
}

// func Check[V comparable](create func(v V) Checker[V], valids []V, invalids []V) {
// 	for i, valid := range valids {
// 		v := create(valid)
// 		// check it is a type of Value
// 	}
// }
