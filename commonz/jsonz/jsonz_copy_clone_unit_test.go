package jsonz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type copyCloneTestStruct struct {
	Name  string
	Value int
	Tags  []string
}

func TestUnitCopy(t *testing.T) {
	type destStruct struct {
		Name  string
		Value int
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
		src := &copyCloneTestStruct{Name: "test", Value: 1, Tags: []string{"a", "b"}}
		dst := &copyCloneTestStruct{}
		check(t, testScenario{
			input:  src,
			output: dst,
			want:   src,
		})
		require.NotSame(t, src, dst)
	})

	t.Run("DifferentStruct", func(t *testing.T) {
		src := &copyCloneTestStruct{Name: "test", Value: 1, Tags: []string{"a", "b"}}
		dst := &destStruct{}
		check(t, testScenario{
			input:  src,
			output: dst,
			want:   &destStruct{Name: "test", Value: 1},
		})
	})

	t.Run("MarshalError", func(t *testing.T) {
		check(t, testScenario{
			input:   make(chan int),
			output:  &copyCloneTestStruct{},
			wantErr: "failed to marshal data",
		})
	})

	t.Run("UnmarshalError-NonPointer", func(t *testing.T) {
		check(t, testScenario{
			input:   &copyCloneTestStruct{Name: "test"},
			output:  copyCloneTestStruct{}, // Non-pointer destination
			wantErr: "failed to parse json",
		})
	})
}

func TestUnitClone(t *testing.T) {
	type testScenario struct {
		name    string
		input   *copyCloneTestStruct
		output  *copyCloneTestStruct
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
		src := &copyCloneTestStruct{Name: "clone test", Value: 123, Tags: []string{"c", "d"}}
		dst := &copyCloneTestStruct{}
		check(t, testScenario{
			input:  src,
			output: dst,
		})
	})

	t.Run("NilInput", func(t *testing.T) {
		dst := &copyCloneTestStruct{}
		check(t, testScenario{
			input:  nil,
			output: dst,
		})
	})

	t.Run("UnmarshalError-NilOutput", func(t *testing.T) {
		check(t, testScenario{
			input:   &copyCloneTestStruct{Name: "test"},
			output:  nil, // nil destination
			wantErr: "failed to parse json",
		})
	})
}
