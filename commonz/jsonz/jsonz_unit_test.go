package jsonz

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"encoding/json"
)

type testStruct struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

type testStructWithNumber struct {
	Value json.Number `json:"value"`
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

	t.Run("json number", func(t *testing.T) {
		data := `{"value": 12345678901234567890}`
		result, err := Parse[testStructWithNumber](data)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, json.Number("12345678901234567890"), result.Value)
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

	t.Run("json number", func(t *testing.T) {
		data := `{"value": 12345678901234567890}`
		require.NotPanics(t, func() {
			result := MustParse[testStructWithNumber](data)
			require.NotNil(t, result)
			require.Equal(t, json.Number("12345678901234567890"), result.Value)
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

func TestUnitNewReader(t *testing.T) {
	type testScenario struct {
		name  string
		items []testStruct
	}

	check := func(t *testing.T, s testScenario) {
		var buf bytes.Buffer
		encoder := json.NewEncoder(&buf)

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

		// After reading all items, next call should return nil, io.EOF
		item, err := reader.ReadItem()
		require.Equal(t, io.EOF, err)
		require.Nil(t, item)
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
		item, err := reader.ReadItem()
		require.Equal(t, io.EOF, err)
		require.Nil(t, item)
	})

	t.Run("invalid json in stream", func(t *testing.T) {
		var buf bytes.Buffer
		buf.WriteString(`invalid json string`) // Write invalid JSON
		reader := NewReader[testStruct](&buf)
		_, err := reader.ReadItem()
		require.Error(t, err)
		require.NotEqual(t, io.EOF, err) // It's an error, not EOF
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
		require.Equal(t, io.EOF, err)
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

func TestUnitNewReaderWriter(t *testing.T) {
	type testScenario struct {
		name  string
		items []*testStruct
	}

	check := func(t *testing.T, s testScenario) {
		var buf bytes.Buffer
		rw := NewReaderWriter[testStruct](&buf)

		for _, item := range s.items {
			err := rw.WriteItem(item)
			require.NoError(t, err)
		}

		// Now read them back to verify
		for i := 0; i < len(s.items); i++ {
			item, err := rw.ReadItem()
			require.NoError(t, err)
			require.NotNil(t, item)
			require.Equal(t, *s.items[i], *item)
		}

		// Check for EOF
		item, err := rw.ReadItem()
		require.Equal(t, io.EOF, err)
		require.Nil(t, item)
	}

	t.Run("read/write multiple items", func(t *testing.T) {
		check(t, testScenario{
			name: "read/write multiple items",
			items: []*testStruct{
				{Foo: "one", Bar: 1},
				{Foo: "two", Bar: 2},
				{Foo: "three", Bar: 3},
			},
		})
	})

	t.Run("read/write single item", func(t *testing.T) {
		check(t, testScenario{
			name: "read/write single item",
			items: []*testStruct{
				{Foo: "single", Bar: 100},
			},
		})
	})

	t.Run("read/write no items", func(t *testing.T) {
		check(t, testScenario{
			name:  "read/write no items",
			items: []*testStruct{},
		})
	})
}

