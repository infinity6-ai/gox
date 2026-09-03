package jsonz

import (
	"bytes"
	"math"
	"testing"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/stretchr/testify/require"
)

type jsonzMustTestStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestUnitJsonzMustParseReader(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		json := `{"name":"test","value":123}`
		reader := bytes.NewReader([]byte(json))
		var result jsonzMustTestStruct
		var parsed *jsonzMustTestStruct

		require.NotPanics(t, func() {
			parsed = MustParseReader(reader, &result)
		})
		require.Equal(t, "test", result.Name)
		require.Equal(t, 123, result.Value)
		require.Equal(t, &result, parsed)
	})

	t.Run("panic on invalid json", func(t *testing.T) {
		json := `{"name":"test","value":123` // invalid json
		reader := bytes.NewReader([]byte(json))
		require.Panics(t, func() {
			MustParseReader(reader, &jsonzMustTestStruct{})
		})
	})
}

func TestUnitJsonzMustParse(t *testing.T) {
	t.Run("success with bytes", func(t *testing.T) {
		json := []byte(`{"name":"test","value":123}`)
		var result jsonzMustTestStruct
		var parsed *jsonzMustTestStruct
		require.NotPanics(t, func() {
			parsed = MustParse(json, &result)
		})
		require.Equal(t, "test", result.Name)
		require.Equal(t, 123, result.Value)
		require.Equal(t, &result, parsed)
	})

	t.Run("success with string", func(t *testing.T) {
		json := `{"name":"test","value":123}`
		var result jsonzMustTestStruct
		var parsed *jsonzMustTestStruct
		require.NotPanics(t, func() {
			parsed = MustParse(json, &result)
		})
		require.Equal(t, "test", result.Name)
		require.Equal(t, 123, result.Value)
		require.Equal(t, &result, parsed)
	})

	t.Run("panic on invalid json", func(t *testing.T) {
		json := `{"name":"test","value":123`
		require.Panics(t, func() {
			MustParse(json, &jsonzMustTestStruct{})
		})
	})
}

func TestUnitJsonzMustFormatWriter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data := jsonzMustTestStruct{Name: "test", Value: 123}
		var buf bytes.Buffer
		require.NotPanics(t, func() {
			MustFormatWriter(&buf, data)
		})
		expectedJson := `{"name":"test","value":123}`
		require.JSONEq(t, expectedJson, buf.String())
	})

	t.Run("panic on unmarshallable type", func(t *testing.T) {
		var buf bytes.Buffer
		// Channels are not marshallable to JSON
		data := make(chan int)
		require.Panics(t, func() {
			MustFormatWriter(&buf, data)
		})
	})
}

func TestUnitJsonzMustFormat(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data := jsonzMustTestStruct{Name: "test", Value: 123}
		var formatted blobz.Blob
		require.NotPanics(t, func() {
			formatted = MustFormat(data)
		})
		expectedJson := `{"name":"test","value":123}`
		require.JSONEq(t, expectedJson, formatted.String())
	})

	t.Run("panic on unmarshallable type", func(t *testing.T) {
		// Functions are not marshallable to JSON
		data := func() {}
		require.Panics(t, func() {
			MustFormat(data)
		})
	})
}

func TestUnitJsonzMustCopy(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		input := &jsonzMustTestStruct{Name: "original", Value: 1}
		var output jsonzMustTestStruct
		var result *jsonzMustTestStruct
		require.NotPanics(t, func() {
			result = MustCopy(input, &output)
		})
		require.Equal(t, *input, output)
		require.Equal(t, &output, result)
		require.NotSame(t, input, &output)
	})

	t.Run("panic on incompatible types", func(t *testing.T) {
		input := &jsonzMustTestStruct{Name: "original", Value: 1}
		// json.Unmarshal will error if you try to unmarshal a struct into an int
		var output int
		require.Panics(t, func() {
			MustCopy(input, &output)
		})
	})
}

func TestUnitJsonzMustClone(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		input := &jsonzMustTestStruct{Name: "original", Value: 1}
		var output jsonzMustTestStruct
		var cloned *jsonzMustTestStruct
		require.NotPanics(t, func() {
			cloned = MustClone(input, &output)
		})
		require.Equal(t, *input, *cloned)
		require.NotSame(t, input, cloned)
	})

	t.Run("panic on unmarshallable type", func(t *testing.T) {
		// math.Inf is not valid in JSON
		require.Panics(t, func() {
			MustClone(math.Inf(1), 0.0)
		})
	})
}
