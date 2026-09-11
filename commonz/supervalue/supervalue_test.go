package supervalue_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/supervalue"
	"github.com/stretchr/testify/require"
)

type MyValue supervalue.SuperValue[string]

func (m MyValue) Check(v string) {
	if v == "" {
		panic("NOOO")
	}
}

func NewMyValue(val string) MyValue {
	ret := MyValue{}
	supervalue.Set(&ret, val)
	return ret
}

func (m MyValue) MarshalJSON() ([]byte, error) {
	return supervalue.Marshal[string](m)
}

func (m *MyValue) UnmarshalJSON(data []byte) error {
	return supervalue.Unmarshal(m, data)
}

func TestUnitBasic(t *testing.T) {
	a1 := NewMyValue("a")
	b1 := NewMyValue("b")
	a2 := NewMyValue("a")

	require.True(t, a1 == a2)
	require.False(t, a1 == b1)

	var x1, x2, x3 MyValue
	jsonz.MustClone(&a1, &x1)
	jsonz.MustClone(&b1, &x2)
	jsonz.MustClone(&a1, &x3)

	require.True(t, x1 == x3)

	require.PanicsWithValue(t, "NOOO", func() {
		NewMyValue("")
	})

	require.PanicsWithValue(t, "x has already been set", func() {
		supervalue.Set(&a1, "a")
	})

	// require.PanicsWithValue(t, "x has already been set", func() {
	// jsonz.MustClone(&a1, &x3)
	// })

	require.Equal(t, `"b"`, jsonz.MustFormat(b1).String())

}
