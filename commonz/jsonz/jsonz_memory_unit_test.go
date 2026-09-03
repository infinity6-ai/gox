package jsonz

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitParseReader(t *testing.T) {
	type MyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	t.Run("Valid JSON string to struct", func(t *testing.T) {
		initial := &MyStruct{}
		want := &MyStruct{Name: "John Doe", Age: 30}
		input := strings.NewReader(`{"name":"John Doe","age":30}`)

		got, err := ParseReader(input, initial)

		require.NoError(t, err)
		require.Equal(t, want, got)
		// The original 'initial' is modified by reference
		require.Equal(t, want, initial)
	})

	t.Run("Valid JSON string to map", func(t *testing.T) {
		initial := &map[string]any{}
		want := &map[string]any{"key": "value", "number": json.Number("123")}
		input := strings.NewReader(`{"key":"value", "number":123}`)

		got, err := ParseReader(input, initial)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("Invalid JSON format", func(t *testing.T) {
		initial := &MyStruct{}
		input := strings.NewReader(`{"key":"value", "number":}`)
		wantErr := "failed to parse json: invalid character '}' looking for beginning of value"

		got, err := ParseReader(input, initial)

		require.Error(t, err)
		require.Contains(t, err.Error(), wantErr)
		require.Equal(t, &MyStruct{}, got) // On error, returns the initial value
	})

	t.Run("Empty input", func(t *testing.T) {
		initial := &MyStruct{}
		input := strings.NewReader(``)
		wantErr := "failed to parse json: EOF"

		got, err := ParseReader(input, initial)
		require.Error(t, err)
		require.Contains(t, err.Error(), wantErr)
		require.Equal(t, &MyStruct{}, got)
	})

	t.Run("Input with numbers is parsed as json.Number", func(t *testing.T) {
		type MyNumStruct struct {
			Value json.Number `json:"value"`
		}
		initial := &MyNumStruct{}
		want := &MyNumStruct{Value: json.Number("123.45")}
		input := strings.NewReader(`{"value":123.45}`)

		got, err := ParseReader(input, initial)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})
}

func TestUnitParse(t *testing.T) {
	type MyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	t.Run("Valid JSON string to struct using Parse", func(t *testing.T) {
		initial := &MyStruct{}
		want := &MyStruct{Name: "Jane Doe", Age: 25}
		input := `{"name":"Jane Doe","age":25}`

		got, err := Parse(input, initial)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("Valid JSON byte slice to map using Parse", func(t *testing.T) {
		initial := &map[string]any{}
		want := &map[string]any{"status": "success"}
		input := []byte(`{"status":"success"}`)

		got, err := Parse(input, initial)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("Invalid JSON string using Parse", func(t *testing.T) {
		initial := &MyStruct{}
		input := `{"status":`
		wantErr := "failed to parse json: unexpected end of JSON input"

		got, err := Parse(input, initial)

		require.Error(t, err)
		require.Contains(t, err.Error(), wantErr)
		require.Equal(t, &MyStruct{}, got)
	})

	t.Run("Empty string using Parse", func(t *testing.T) {
		initial := &MyStruct{}
		input := ""
		wantErr := "failed to parse json: EOF"

		got, err := Parse(input, initial)
		require.Error(t, err)
		require.Contains(t, err.Error(), wantErr)
		require.Equal(t, &MyStruct{}, got)
	})
}

func TestUnitFormatWriter(t *testing.T) {
	type MyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	type testScenario struct {
		name    string
		input   any
		want    string
		wantErr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		buf := new(bytes.Buffer)
		err := FormatWriter(buf, s.input)

		if s.wantErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.wantErr)
		} else {
			require.NoError(t, err)
			require.JSONEq(t, s.want, buf.String())
		}
	}

	t.Run("Format struct to JSON writer", func(t *testing.T) {
		input := MyStruct{Name: "Alice", Age: 40}
		check(t, testScenario{
			name:    "format struct",
			input:   input,
			want:    `{"name":"Alice","age":40}`,
			wantErr: "",
		})
	})

	t.Run("Format map to JSON writer", func(t *testing.T) {
		input := map[string]any{"city": "New York", "population": 8000000}
		check(t, testScenario{
			name:    "format map",
			input:   input,
			want:    `{"city":"New York","population":8000000}`,
			wantErr: "",
		})
	})

	t.Run("Format nil input", func(t *testing.T) {
		check(t, testScenario{
			name:    "format nil",
			input:   nil,
			want:    `null`,
			wantErr: "",
		})
	})

	t.Run("Format unsupported type (channel) to JSON writer", func(t *testing.T) {
		input := make(chan int)
		check(t, testScenario{
			name:    "format unsupported type",
			input:   input,
			want:    "",
			wantErr: "failed to marshal data: json: unsupported type: chan int",
		})
	})
}

func TestUnitFormat(t *testing.T) {
	type MyStruct struct {
		Product string  `json:"product"`
		Price   float64 `json:"price"`
	}

	type testScenario struct {
		name    string
		input   any
		want    string
		wantErr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got, err := Format(s.input)

		if s.wantErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.wantErr)
			require.Nil(t, got)
		} else {
			require.NoError(t, err)
			require.JSONEq(t, s.want, got.String())
		}
	}

	t.Run("Format struct to blob", func(t *testing.T) {
		input := MyStruct{Product: "Laptop", Price: 1200.50}
		check(t, testScenario{
			name:    "format struct",
			input:   input,
			want:    `{"product":"Laptop","price":1200.5}`,
			wantErr: "",
		})
	})

	t.Run("Format map to blob", func(t *testing.T) {
		input := map[string]any{"item": "Book", "quantity": 1}
		check(t, testScenario{
			name:    "format map",
			input:   input,
			want:    `{"item":"Book","quantity":1}`,
			wantErr: "",
		})
	})

	t.Run("Format slice to blob", func(t *testing.T) {
		input := []int{1, 2, 3}
		check(t, testScenario{
			name:    "format slice",
			input:   input,
			want:    `[1,2,3]`,
			wantErr: "",
		})
	})

	t.Run("Format unsupported type (function) to blob", func(t *testing.T) {
		input := func() {}
		check(t, testScenario{
			name:    "format unsupported type",
			input:   input,
			want:    "",
			wantErr: "failed to marshal data: json: unsupported type: func()",
		})
	})
}
