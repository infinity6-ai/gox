package gobz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Foo string
	Bar int
}

func TestUnitParse(t *testing.T) {
	t.Run("valid gob", func(t *testing.T) {
		// First, create a gob from a struct
		input := &testStruct{Foo: "hello", Bar: 123}
		data, err := Format(input)
		require.NoError(t, err)

		// Now, parse it back
		result, err := Parse[testStruct](data)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "hello", result.Foo)
		require.Equal(t, 123, result.Bar)
	})

	t.Run("invalid gob", func(t *testing.T) {
		data := []byte("this is not gob")
		_, err := Parse[testStruct](data)
		require.Error(t, err)
	})
}

func TestUnitMustParse(t *testing.T) {
	t.Run("valid gob", func(t *testing.T) {
		input := &testStruct{Foo: "hello", Bar: 123}
		data, err := Format(input)
		require.NoError(t, err)

		require.NotPanics(t, func() {
			result := MustParse[testStruct](data)
			require.NotNil(t, result)
			require.Equal(t, "hello", result.Foo)
			require.Equal(t, 123, result.Bar)
		})
	})

	t.Run("invalid gob", func(t *testing.T) {
		data := []byte("this is not gob")
		require.Panics(t, func() {
			MustParse[testStruct](data)
		})
	})
}

func TestUnitFormat(t *testing.T) {
	data := &testStruct{Foo: "hello", Bar: 123}
	bytes, err := Format(data)
	require.NoError(t, err)
	require.NotNil(t, bytes)

	// Let's parse it back to be sure.
	result, err := Parse[testStruct](bytes)
	require.NoError(t, err)
	require.Equal(t, data, result)
}

func TestUnitMustFormat(t *testing.T) {
	t.Run("valid struct", func(t *testing.T) {
		data := &testStruct{Foo: "hello", Bar: 123}
		require.NotPanics(t, func() {
			bytes := MustFormat(data)
			require.NotNil(t, bytes)
			result, err := Parse[testStruct](bytes)
			require.NoError(t, err)
			require.Equal(t, data, result)
		})
	})
}
