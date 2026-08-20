package jsonz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

func TestUnitParse(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		data := `{"foo": "hello", "bar": 123}`
		result, err := Parse[testStruct](data)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "hello", result.Foo)
		require.Equal(t, 123, result.Bar)
	})

	t.Run("bytes", func(t *testing.T) {
		data := []byte(`{"foo": "world", "bar": 456}`)
		result, err := Parse[testStruct](data)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "world", result.Foo)
		require.Equal(t, 456, result.Bar)
	})

	t.Run("invalid json", func(t *testing.T) {
		data := `{"foo": "hello", "bar": 123`
		_, err := Parse[testStruct](data)
		require.Error(t, err)
	})
}

func TestUnitMustParse(t *testing.T) {
	t.Run("valid json", func(t *testing.T) {
		data := `{"foo": "hello", "bar": 123}`
		require.NotPanics(t, func() {
			result := MustParse[testStruct](data)
			require.NotNil(t, result)
			require.Equal(t, "hello", result.Foo)
			require.Equal(t, 123, result.Bar)
		})
	})

	t.Run("invalid json", func(t *testing.T) {
		data := `{"foo": "hello", "bar": 123`
		require.Panics(t, func() {
			MustParse[testStruct](data)
		})
	})
}

func TestUnitFormat(t *testing.T) {
	data := &testStruct{Foo: "hello", Bar: 123}
	blob, err := Format(data)
	require.NoError(t, err)
	require.NotNil(t, blob)
	// Comparing strings for JSON is tricky due to key order and whitespace.
	// Let's parse it back to be sure.
	result, err := Parse[testStruct](blob.String())
	require.NoError(t, err)
	require.Equal(t, data, result)
}

func TestUnitMustFormat(t *testing.T) {
	t.Run("valid struct", func(t *testing.T) {
		data := &testStruct{Foo: "hello", Bar: 123}
		require.NotPanics(t, func() {
			blob := MustFormat(data)
			require.NotNil(t, blob)
			result, err := Parse[testStruct](blob.String())
			require.NoError(t, err)
			require.Equal(t, data, result)
		})
	})
}
