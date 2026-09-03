package legacyjsonz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type sampleStruct struct {
	Name  string
	Value int
	Tags  []string
}

func TestUnitCopy(t *testing.T) {
	type testScenario struct {
		name       string
		input      sampleStruct
		output     *sampleStruct
		wantOutput sampleStruct
		wantErr    bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		res, err := Copy(s.input, s.output)
		if s.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, s.wantOutput, *s.output)
			require.Same(t, s.output, res)
		}
	}

	t.Run("Successful copy of struct", func(t *testing.T) {
		input := sampleStruct{Name: "test1", Value: 10, Tags: []string{"a", "b"}}
		var output sampleStruct
		check(t, testScenario{
			name:       "Successful copy",
			input:      input,
			output:     &output,
			wantOutput: input,
			wantErr:    false,
		})
	})

	t.Run("Copy to nil output should error", func(t *testing.T) {
		input := sampleStruct{Name: "test2", Value: 20}
		check(t, testScenario{
			name:    "Copy to nil output",
			input:   input,
			output:  nil, // This will cause an error during Unmarshal
			wantErr: true,
		})
	})
}

func TestUnitMustCopy(t *testing.T) {
	type testScenario struct {
		name       string
		input      sampleStruct
		output     *sampleStruct
		wantOutput sampleStruct
		wantPanic  bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.wantPanic {
			require.Panics(t, func() {
				MustCopy(s.input, s.output)
			})
		} else {
			var res *sampleStruct
			require.NotPanics(t, func() {
				res = MustCopy(s.input, s.output)
			})
			require.Equal(t, s.wantOutput, *s.output)
			require.Same(t, s.output, res)
		}
	}

	t.Run("Successful MustCopy of struct", func(t *testing.T) {
		input := sampleStruct{Name: "test3", Value: 30, Tags: []string{"c"}}
		var output sampleStruct
		check(t, testScenario{
			name:       "Successful MustCopy",
			input:      input,
			output:     &output,
			wantOutput: input,
			wantPanic:  false,
		})
	})

	t.Run("MustCopy with nil output should panic", func(t *testing.T) {
		input := sampleStruct{Name: "test4", Value: 40}
		check(t, testScenario{
			name:      "MustCopy with nil output",
			input:     input,
			output:    nil, // This will cause an error during Unmarshal and then panic
			wantPanic: true,
		})
	})
}

func TestUnitClone(t *testing.T) {
	type testScenario struct {
		name      string
		input     sampleStruct
		wantClone sampleStruct
		wantErr   bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		cloned, err := Clone(s.input)
		if s.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, s.wantClone, cloned)
			// Ensure it's a deep copy by modifying input and checking clone
			s.input.Value = 999
			require.NotEqual(t, s.input.Value, cloned.Value)
		}
	}

	t.Run("Successful clone of struct", func(t *testing.T) {
		input := sampleStruct{Name: "original", Value: 100, Tags: []string{"x", "y"}}
		check(t, testScenario{
			name:      "Successful clone",
			input:     input,
			wantClone: input,
			wantErr:   false,
		})
	})
}

func TestUnitMustClone(t *testing.T) {
	type testScenario struct {
		name      string
		input     sampleStruct
		wantClone sampleStruct
		wantPanic bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.wantPanic {
			require.Panics(t, func() {
				MustClone(s.input)
			})
		} else {
			var cloned sampleStruct
			require.NotPanics(t, func() {
				cloned = MustClone(s.input)
			})
			require.Equal(t, s.wantClone, cloned)
			// Ensure it's a deep copy by modifying input and checking clone
			s.input.Value = 999
			require.NotEqual(t, s.input.Value, cloned.Value)
		}
	}

	t.Run("Successful MustClone of struct", func(t *testing.T) {
		input := sampleStruct{Name: "originalMust", Value: 200, Tags: []string{"m", "n"}}
		check(t, testScenario{
			name:      "Successful MustClone",
			input:     input,
			wantClone: input,
			wantPanic: false,
		})
	})
}
