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
		var result testStruct
		err = Parse[testStruct](data, &result)
		require.NoError(t, err)
		require.Equal(t, "hello", result.Foo)
		require.Equal(t, 123, result.Bar)
	})

	t.Run("invalid gob", func(t *testing.T) {
		data := []byte("this is not gob")
		var result testStruct
		err := Parse[testStruct](data, &result)
		require.Error(t, err)
	})
}

func TestUnitMustParse(t *testing.T) {
	t.Run("valid gob", func(t *testing.T) {
		input := &testStruct{Foo: "hello", Bar: 123}
		data, err := Format(input)
		require.NoError(t, err)

		require.NotPanics(t, func() {
			var result testStruct
			MustParse[testStruct](data, &result)
			require.Equal(t, "hello", result.Foo)
			require.Equal(t, 123, result.Bar)
		})
	})

	t.Run("invalid gob", func(t *testing.T) {
		data := []byte("this is not gob")
		require.Panics(t, func() {
			var result testStruct
			MustParse[testStruct](data, &result)
		})
	})
}

func TestUnitFormat(t *testing.T) {
	data := &testStruct{Foo: "hello", Bar: 123}
	bytes, err := Format(data)
	require.NoError(t, err)
	require.NotNil(t, bytes)

	// Let's parse it back to be sure.
	var result *testStruct
	err = Parse[*testStruct](bytes, &result)
	require.NoError(t, err)
	require.Equal(t, data, result)
}

func TestUnitMustFormat(t *testing.T) {
	t.Run("valid struct", func(t *testing.T) {
		data := &testStruct{Foo: "hello", Bar: 123}
		require.NotPanics(t, func() {
			bytes := MustFormat(data)
			require.NotNil(t, bytes)
			var result *testStruct
			err := Parse[*testStruct](bytes, &result)
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
			require.Equal(t, s.items[i].Foo, item.Foo)
			require.Equal(t, s.items[i].Bar, item.Bar)
		}

		// After reading all items, next call should return a zero value and io.EOF
		item, err := reader.ReadItem()
		require.Equal(t, io.EOF, err)
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
		require.Equal(t, io.EOF, err)
		require.Equal(t, testStruct{}, item)
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
		require.Equal(t, io.EOF, err)
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
		require.Equal(t, io.EOF, err)
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
