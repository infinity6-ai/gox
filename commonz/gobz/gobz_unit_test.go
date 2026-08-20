package gobz

import (
	"bytes"
	"encoding/gob"
	"io"
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

func TestUnitNewReader(t *testing.T) {
	type testScenario struct {
		name  string
		items []testStruct
	}

	check := func(t *testing.T, s testScenario) {
		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)

		for _, item := range s.items {
			err := encoder.Encode(item)
			require.NoError(t, err)
		}

		reader := NewReader[testStruct](&buf)

		for i := 0; i < len(s.items); i++ {
			item, err := reader.ReadItem()
			require.NoError(t, err)
			require.NotNil(t, item)
			require.Equal(t, s.items[i].Foo, item.Foo)
			require.Equal(t, s.items[i].Bar, item.Bar)
		}

		// After reading all items, next call should return nil, nil
		item, err := reader.ReadItem()
		require.NoError(t, err) // Expect no error for EOF
		require.Nil(t, item)    // Expect nil item for EOF
	}

	t.Run("read multiple items", func(t *testing.T) {
		check(t, testScenario{
			name: "read multiple items",
			items: []testStruct{
				{Foo: "one", Bar: 1},
				{Foo: "two", Bar: 2},
				{Foo: "three", Bar: 3},
			},
		})
	})

	t.Run("read single item", func(t *testing.T) {
		check(t, testScenario{
			name: "read single item",
			items: []testStruct{
				{Foo: "single", Bar: 100},
			},
		})
	})

	t.Run("read empty stream", func(t *testing.T) {
		var buf bytes.Buffer
		reader := NewReader[testStruct](&buf)
		item, err := reader.ReadItem() // Expect nil, nil for EOF
		require.NoError(t, err)
		require.Nil(t, item)
	})

	t.Run("invalid gob in stream", func(t *testing.T) {
		var buf bytes.Buffer
		buf.Write([]byte("this is not a gob"))

		reader := NewReader[testStruct](&buf)
		_, err := reader.ReadItem()
		require.Error(t, err)
		// The error should not be io.EOF here, as it's an invalid gob, not just end of stream.
		require.NotEqual(t, io.EOF, err)
	})
}

func TestUnitNewWriter(t *testing.T) {
	type testScenario struct {
		name  string
		items []*testStruct
	}

	check := func(t *testing.T, s testScenario) {
		var buf bytes.Buffer
		writer := NewWriter[testStruct](&buf)

		for _, item := range s.items {
			err := writer.WriteItem(item)
			require.NoError(t, err)
		}

		// Now read them back to verify
		reader := NewReader[testStruct](&buf)
		for i := 0; i < len(s.items); i++ {
			item, err := reader.ReadItem()
			require.NoError(t, err)
			require.NotNil(t, item)
			require.Equal(t, *s.items[i], *item)
		}

		// Check for EOF
		item, err := reader.ReadItem()
		require.NoError(t, err)
		require.Nil(t, item)
	}

	t.Run("write multiple items", func(t *testing.T) {
		check(t, testScenario{
			name: "write multiple items",
			items: []*testStruct{
				{Foo: "one", Bar: 1},
				{Foo: "two", Bar: 2},
				{Foo: "three", Bar: 3},
			},
		})
	})

	t.Run("write single item", func(t *testing.T) {
		check(t, testScenario{
			name: "write single item",
			items: []*testStruct{
				{Foo: "single", Bar: 100},
			},
		})
	})

	t.Run("write no items", func(t *testing.T) {
		check(t, testScenario{
			name:  "write no items",
			items: []*testStruct{},
		})
	})
}
