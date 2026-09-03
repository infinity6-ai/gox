package jsonz

import (
	"bytes"
	"io"
	"testing"

	"encoding/json"

	"github.com/stretchr/testify/require"
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
		var result testStruct
		err := Parse(data, &result)
		require.NoError(t, err)
		require.Equal(t, "hello", result.Foo)
		require.Equal(t, 123, result.Bar)
	})

	t.Run("bytes", func(t *testing.T) {
		data := []byte(`{"foo": "world", "bar": 456}`)
		var result testStruct
		err := Parse(data, &result)
		require.NoError(t, err)
		require.Equal(t, "world", result.Foo)
		require.Equal(t, 456, result.Bar)
	})

	t.Run("invalid json", func(t *testing.T) {
		data := `{"foo": "hello", "bar": 123`
		var result testStruct
		err := Parse(data, &result)
		require.Error(t, err)
	})

	t.Run("json number", func(t *testing.T) {
		data := `{"value": 12345678901234567890}`
		var result testStructWithNumber
		err := Parse(data, &result)
		require.NoError(t, err)
		require.Equal(t, json.Number("12345678901234567890"), result.Value)
	})

	t.Run("nil out reference", func(t *testing.T) {
		data := `{"foo": "hello", "bar": 123}`
		err := Parse[testStruct](data, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "output reference cannot be nil")
	})
}

func TestUnitMustParse(t *testing.T) {
	t.Run("valid json", func(t *testing.T) {
		data := `{"foo": "hello", "bar": 123}`
		var result testStruct
		require.NotPanics(t, func() {
			MustParse(data, &result)
		})
		require.Equal(t, "hello", result.Foo)
		require.Equal(t, 123, result.Bar)
	})

	t.Run("invalid json", func(t *testing.T) {
		data := `{"foo": "hello", "bar": 123`
		var result testStruct
		require.Panics(t, func() {
			MustParse(data, &result)
		})
	})

	t.Run("json number", func(t *testing.T) {
		data := `{"value": 12345678901234567890}`
		var result testStructWithNumber
		require.NotPanics(t, func() {
			MustParse(data, &result)
		})
		require.Equal(t, json.Number("12345678901234567890"), result.Value)
	})

	t.Run("nil out reference", func(t *testing.T) {
		data := `{"foo": "hello", "bar": 123}`
		require.Panics(t, func() {
			MustParse[testStruct](data, nil)
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
	var result testStruct
	err = Parse(blob.String(), &result)
	require.NoError(t, err)
	require.Equal(t, *data, result)
}

func TestUnitMustFormat(t *testing.T) {
	t.Run("valid struct", func(t *testing.T) {
		data := &testStruct{Foo: "hello", Bar: 123}
		require.NotPanics(t, func() {
			blob := MustFormat(data)
			require.NotNil(t, blob)
			var result testStruct
			err := Parse(blob.String(), &result)
			require.NoError(t, err)
			require.Equal(t, *data, result)
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
			require.Equal(t, s.items[i].Foo, item.Foo)
			require.Equal(t, s.items[i].Bar, item.Bar)
		}

		// After reading all items, next call should return a zero value and io.EOF
		item, err := reader.ReadItem()
		require.ErrorIs(t, err, io.EOF)
		require.Equal(t, testStruct{}, item)
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
		require.ErrorIs(t, err, io.EOF)
		require.Equal(t, testStruct{}, item)
	})

	t.Run("invalid json in stream", func(t *testing.T) {
		var buf bytes.Buffer
		buf.WriteString(`invalid json string`) // Write invalid JSON
		reader := NewReader[testStruct](&buf)
		_, err := reader.ReadItem()
		require.Error(t, err)
		require.NotErrorIs(t, err, io.EOF) // It's an error, not EOF
	})
}

func TestUnitNewWriter(t *testing.T) {
	type testScenario struct {
		name  string
		items []testStruct
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
			require.Equal(t, s.items[i], item)
		}

		// Check for EOF
		item, err := reader.ReadItem()
		require.ErrorIs(t, err, io.EOF)
		require.Equal(t, testStruct{}, item)
	}

	t.Run("write multiple items", func(t *testing.T) {
		check(t, testScenario{
			name: "write multiple items",
			items: []testStruct{
				{Foo: "one", Bar: 1},
				{Foo: "two", Bar: 2},
				{Foo: "three", Bar: 3},
			},
		})
	})

	t.Run("write single item", func(t *testing.T) {
		check(t, testScenario{
			name: "write single item",
			items: []testStruct{
				{Foo: "single", Bar: 100},
			},
		})
	})

	t.Run("write no items", func(t *testing.T) {
		check(t, testScenario{
			name:  "write no items",
			items: []testStruct{},
		})
	})
}

func TestUnitNewReaderWriter(t *testing.T) {
	type testScenario struct {
		name  string
		items []testStruct
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
			require.Equal(t, s.items[i], item)
		}

		// Check for EOF
		item, err := rw.ReadItem()
		require.ErrorIs(t, err, io.EOF)
		require.Equal(t, testStruct{}, item)
	}

	t.Run("read/write multiple items", func(t *testing.T) {
		check(t, testScenario{
			name: "read/write multiple items",
			items: []testStruct{
				{Foo: "one", Bar: 1},
				{Foo: "two", Bar: 2},
				{Foo: "three", Bar: 3},
			},
		})
	})

	t.Run("read/write single item", func(t *testing.T) {
		check(t, testScenario{
			name: "read/write single item",
			items: []testStruct{
				{Foo: "single", Bar: 100},
			},
		})
	})

	t.Run("read/write no items", func(t *testing.T) {
		check(t, testScenario{
			name:  "read/write no items",
			items: []testStruct{},
		})
	})
}
