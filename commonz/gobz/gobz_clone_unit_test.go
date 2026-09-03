package gobz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
