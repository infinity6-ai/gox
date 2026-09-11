package supervalue_test

import (
	"testing"

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

}
