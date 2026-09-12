package tuchecked

import (
	"reflect"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/validation/checked"
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
		var first T
		err := errorz.UnpanicV(func() {
			first = table.Create(valid)
		})
		checker.Nil(err, "panic while creating [idx=%d]: %v", idx, valid)
		second := table.Create(valid)
		checker.True(reflect.ValueOf(first).Elem().Equal(reflect.ValueOf(second).Elem()), "creating with same value must be equal, expected: %v, but was: %v", valid, valid)
		cvFirst := any(first).(checked.Checker[V])
		checker.Equal(cvFirst.Get(), valid, "Get method must be equal: %v", valid)
		cvSecond := any(second).(checked.Checker[V])
		checker.Equal(cvSecond.Get(), valid, "Get method must be equal: %v", valid)
		checker.Equal(cvFirst.String(), cvSecond.String(), "String method: %v", valid)

		cvJsonFirst := jsonz.MustFormat(cvFirst).String()
		cvJsonSecond := jsonz.MustFormat(cvSecond).String()
		checker.Equal(cvJsonFirst, cvJsonSecond, "json format: %v", valid)

		// var cvParsedFirst checked.Checker[V]
		// jsonz.MustParse(cvJsonFirst, &cvParsedFirst)
		// checker.True(cvFirst == cvParsedFirst, "json parser: %v, expected: %v, but was: %v", valid, cvFirst, cvParsedFirst)
	}
}

// func Check[V comparable](create func(v V) Checker[V], valids []V, invalids []V) {
// 	for i, valid := range valids {
// 		v := create(valid)
// 		// check it is a type of Value
// 	}
// }
