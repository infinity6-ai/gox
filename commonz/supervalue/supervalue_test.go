package supervalue_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/supervalue"
	"github.com/stretchr/testify/require"
)

// type MyValue supervalue.SuperValue[string]

// func (m MyValue) Check() {

// }

// func NewMyValue(val string) MyValue {
// 	ret := MyValue{}
// 	supervalue.Set(&ret, val)
// 	return ret
// }

func TestUnitBasic(t *testing.T) {
	a1 := supervalue.NewMyValue("a")
	b1 := supervalue.NewMyValue("b")
	a2 := supervalue.NewMyValue("a")

	require.True(t, a1 == a2)
	require.False(t, a1 == b1)

	var x1, x2, x3 supervalue.MyValue
	jsonz.MustClone(&a1, &x1)
	jsonz.MustClone(&b1, &x2)
	jsonz.MustClone(&a1, &x3)

	require.True(t, x1 == x3)

	require.Panics(t, func() {
		supervalue.NewMyValue("c")
	})

	require.Equal(t, `"b"`, jsonz.MustFormat(b1).String())

	// x := supervalue.SuperValue[string](a1)
	// supervalue.Set(a1, "c")

}
