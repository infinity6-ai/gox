package supervalue_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/supervalue"
	"github.com/infinity6-ai/gox/commonz/validation"
	"github.com/stretchr/testify/require"
)

type Dataset supervalue.SuperValue[string]

func (m Dataset) Validate(v string) error {
	return validation.StrNotEmpty(v, "MyValue cannot be empty")
}

func NewMyValue(val string) Dataset {
	ret := Dataset{}
	err := supervalue.Set(&ret, val)
	errorz.Check(err)
	return ret
}

func (m Dataset) MarshalJSON() ([]byte, error) {
	return supervalue.Marshal[string](m)
}

func (m *Dataset) UnmarshalJSON(data []byte) error {
	return supervalue.Unmarshal(m, data)
}

func TestUnitBasic(t *testing.T) {
	a1 := NewMyValue("a")
	b1 := NewMyValue("b")
	a2 := NewMyValue("a")

	require.True(t, a1 == a2)
	require.False(t, a1 == b1)

	var x1, x2, x3 Dataset
	jsonz.MustClone(&a1, &x1)
	jsonz.MustClone(&b1, &x2)
	jsonz.MustClone(&a1, &x3)

	require.True(t, x1 == x3)

	require.PanicsWithError(t, "supervalue validation error: validation error must not be empty: MyValue cannot be empty: (InternalError, code=500)", func() {
		NewMyValue("")
	})

	require.PanicsWithValue(t, "x has already been set", func() {
		supervalue.Set(&a1, "a")
	})

	require.PanicsWithValue(t, "x has already been set", func() {
		jsonz.Clone(&a1, &x3)
	})

	require.Equal(t, `"b"`, jsonz.MustFormat(b1).String())

}
