package checked_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/validation"
	"github.com/infinity6-ai/gox/commonz/validation/checked"
	"github.com/stretchr/testify/require"
)

type MyString checked.Value[string]

func (m MyString) Validate(v string) error {
	return validation.StrNotEmpty(v, "MyValue cannot be empty")
}

func (m MyString) Get() string {
	return checked.Value[string](m).Get()
}

func (m MyString) String() string {
	return m.Get()
}

func NewMyValue(val string) MyString {
	ret := MyString{}
	err := checked.Set(&ret, val)
	errorz.Check(err)
	return ret
}

func (m MyString) MarshalJSON() ([]byte, error) {
	return checked.Marshal[string](m)
}

func (m *MyString) UnmarshalJSON(data []byte) error {
	return checked.Unmarshal(m, data)
}

func TestUnitMyValue(t *testing.T) {
	a1 := NewMyValue("a")
	b1 := NewMyValue("b")
	a2 := NewMyValue("a")

	require.Equal(t, "a", a1.Get())
	require.Equal(t, "b", b1.Get())
	require.Equal(t, "a", a2.Get())

	require.Equal(t, "a", a1.String())
	require.Equal(t, "b", b1.String())
	require.Equal(t, "a", a2.String())

	require.True(t, a1 == a2)
	require.False(t, a1 == b1)

	var x1, x2, x3 MyString
	jsonz.MustClone(&a1, &x1)
	jsonz.MustClone(&b1, &x2)
	jsonz.MustClone(&a1, &x3)

	require.True(t, x1 == x3)

	require.PanicsWithError(t, "supervalue validation error: validation error must not be empty: MyValue cannot be empty: (InternalError, code=500)", func() {
		NewMyValue("")
	})

	require.PanicsWithValue(t, "x has already been set", func() {
		checked.Set(&a1, "a")
	})

	require.PanicsWithValue(t, "x has already been set", func() {
		jsonz.Clone(&a1, &x3)
	})

	require.Equal(t, `"b"`, jsonz.MustFormat(b1).String())

}
