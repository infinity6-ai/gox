package gobz

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitParseReaderAndFormatWriter(t *testing.T) {
	type testData struct {
		Name  string
		Value int
	}

	type testScenario struct {
		name    string
		data    testData
		wantErr bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		var buf bytes.Buffer
		err := FormatWriter(&buf, s.data)
		if s.wantErr {
			require.Error(t, err)
			return
		}
		require.NoError(t, err)

		initialData := &testData{} // Create a non-nil pointer for decoding
		resultGot, err := ParseReader(&buf, initialData)
		if s.wantErr {
			require.Error(t, err)
			return
		}
		require.NoError(t, err)
		require.Equal(t, &s.data, resultGot)


	}

	t.Run("Valid data", func(t *testing.T) {
		check(t, testScenario{
			data: testData{
				Name:  "test",
				Value: 123,
			},
		})
	})

	t.Run("Empty data", func(t *testing.T) {
		check(t, testScenario{
			data: testData{},
		})
	})
}

func TestUnitParse(t *testing.T) {
	type MyStruct struct {
		Name string
		Age  int
	}

	type testScenario[T any] struct {
		name    string
		input   []byte
		initial T
		want    T
		wantErr string
	}

	check := func(t *testing.T, s testScenario[*MyStruct]) {
		t.Helper()
		got, err := Parse(s.input, s.initial)
		if s.wantErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.wantErr)
			require.Equal(t, s.initial, got) // On error, returns the initial value
		} else {
			require.NoError(t, err)
			require.Equal(t, s.want, got)
		}
	}

	t.Run("Valid gob-encoded data to struct", func(t *testing.T) {
		initial := &MyStruct{}
		want := &MyStruct{Name: "John Doe", Age: 30}
		encoded, err := Format(want)
		require.NoError(t, err)

		check(t, testScenario[*MyStruct]{
			input:   encoded,
			initial: initial,
			want:    want,
		})
	})

	t.Run("Invalid gob-encoded data", func(t *testing.T) {
		initial := &MyStruct{}
		input := []byte("invalid gob data")
		wantErr := "failed to gob decode: unexpected EOF"

		check(t, testScenario[*MyStruct]{
			input:   input,
			initial: initial,
			want:    initial, // On error, initial value should be returned
			wantErr: wantErr,
		})
	})

	t.Run("Empty input", func(t *testing.T) {
		initial := &MyStruct{}
		input := []byte("")
		wantErr := "failed to gob decode: EOF"

		check(t, testScenario[*MyStruct]{
			input:   input,
			initial: initial,
			want:    initial, // On error, initial value should be returned
			wantErr: wantErr,
		})
	})
}

func TestUnitParseReader(t *testing.T) {
	type MyStruct struct {
		Name string
		Age  int
	}

	type testScenario[T any] struct {
		name    string
		input   io.Reader
		initial T
		want    T
		wantErr string
	}

	check := func(t *testing.T, s testScenario[*MyStruct]) {
		t.Helper()
		got, err := ParseReader(s.input, s.initial)
		if s.wantErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.wantErr)
			require.Equal(t, s.initial, got) // On error, returns the initial value
		} else {
			require.NoError(t, err)
			require.Equal(t, s.want, got)
		}
	}

	t.Run("Valid gob-encoded reader to struct", func(t *testing.T) {
		initial := &MyStruct{}
		want := &MyStruct{Name: "Jane Doe", Age: 25}
		encoded, err := Format(want)
		require.NoError(t, err)
		input := bytes.NewReader(encoded)

		check(t, testScenario[*MyStruct]{
			input:   input,
			initial: initial,
			want:    want,
		})
	})

	t.Run("Invalid gob-encoded reader", func(t *testing.T) {
		initial := &MyStruct{}
		input := strings.NewReader("invalid gob data")
		wantErr := "failed to gob decode: unexpected EOF"

		check(t, testScenario[*MyStruct]{
			input:   input,
			initial: initial,
			want:    initial, // On error, initial value should be returned
			wantErr: wantErr,
		})
	})

	t.Run("Empty reader", func(t *testing.T) {
		initial := &MyStruct{}
		input := bytes.NewReader([]byte(""))
		wantErr := "failed to gob decode: EOF"

		check(t, testScenario[*MyStruct]{
			input:   input,
			initial: initial,
			want:    initial, // On error, initial value should be returned
			wantErr: wantErr,
		})
	})

	t.Run("Reader returning error", func(t *testing.T) {
		initial := &MyStruct{}
		errReader := &errorReader{
			Reader: strings.NewReader("some data"),
			err:    errors.New("read error"),
		}
		wantErr := "failed to gob decode: read error"

		check(t, testScenario[*MyStruct]{
			input:   errReader,
			initial: initial,
			want:    initial,
			wantErr: wantErr,
		})
	})
}

// errorReader is a helper for testing io.Reader errors
type errorReader struct {
	io.Reader
	err error
}

func (er *errorReader) Read(p []byte) (n int, err error) {
	if er.err != nil {
		return 0, er.err
	}
	return er.Reader.Read(p)
}

func TestUnitCopy(t *testing.T) {
	type testStruct struct {
		Name  string
		Value int
	}
	type destStruct struct {
		Name string
	}

	type testScenario struct {
		name    string
		input   any
		output  any
		want    any
		wantErr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got, err := Copy(s.input, s.output)
		if s.wantErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.wantErr)
			return
		}
		require.NoError(t, err)
		require.Equal(t, s.want, got)
	}

	t.Run("Success", func(t *testing.T) {
		src := &testStruct{Name: "test", Value: 1}
		dst := &testStruct{}
		check(t, testScenario{
			input:  src,
			output: dst,
			want:   src,
		})
		require.NotSame(t, src, dst)
	})

	t.Run("DifferentStruct", func(t *testing.T) {
		check(t, testScenario{
			input:   &testStruct{Name: "test", Value: 1},
			output:  &destStruct{},
			want:    &destStruct{Name: "test"},
			wantErr: "",
		})
	})

	t.Run("EncodeError", func(t *testing.T) {
		check(t, testScenario{
			input:   make(chan int),
			output:  &testStruct{},
			wantErr: "failed to gob encode",
		})
	})

	t.Run("DecodeError-NonPointer", func(t *testing.T) {
		check(t, testScenario{
			input:   &testStruct{Name: "test"},
			output:  testStruct{}, // Non-pointer destination
			wantErr: "failed to gob decode",
		})
	})
}

func TestUnitClone(t *testing.T) {
	type testStruct struct {
		Name  string
		Value int
	}
	type testScenario struct {
		name    string
		input   *testStruct
		output  *testStruct
		wantErr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got, err := Clone(s.input, s.output)
		if s.wantErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.wantErr)
			return
		}
		require.NoError(t, err)
		require.Equal(t, s.input, got)
		if s.input != nil {
			require.NotSame(t, s.input, got)
		}
	}

	t.Run("Success", func(t *testing.T) {
		src := &testStruct{Name: "clone test", Value: 123}
		dst := &testStruct{}
		check(t, testScenario{
			input:  src,
			output: dst,
		})
	})

	t.Run("NilInput", func(t *testing.T) {
		dst := &testStruct{}
		check(t, testScenario{
			input:  nil,
			output: dst,
		})
	})

	t.Run("DecodeError-NilOutput", func(t *testing.T) {
		check(t, testScenario{
			input:   &testStruct{Name: "test"},
			output:  nil, // nil destination
			wantErr: "failed to gob decode",
		})
	})
}
