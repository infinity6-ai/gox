package parserz

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

func TestUnitMustItemReader(t *testing.T) {
	type testScenario[T any] struct {
		name        string
		readFunc    func() (T, error)
		expectPanic bool
		expectedErr string // Used when expectPanic is true
		expectedVal T      // Used when expectPanic is false
	}

	check := func(t *testing.T, scenario testScenario[int]) {
		t.Helper()
		mustReader := NewMustItemReaderWriter(scenario.readFunc, nil)

		if scenario.expectPanic {
			var recovered any
			func() {
				defer func() {
					recovered = recover()
				}()
				mustReader.MustReadItem()
			}()

			require.NotNil(t, recovered, "MustReadItem should have panicked")
			se, ok := recovered.(errorz.StructuredError)
			require.True(t, ok, "Panic value should be a StructuredError")
			require.Contains(t, se.Unwrap().Error(), scenario.expectedErr, "StructuredError should wrap the expected error")
		} else {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("MustReadItem should not have panicked, but panicked with: %v", r)
				}
			}()
			val := mustReader.MustReadItem()
			require.Equal(t, scenario.expectedVal, val, "MustReadItem should return the expected value")
		}
	}

	t.Run("Successfully read item", func(t *testing.T) {
		check(t, testScenario[int]{
			name: "successful read",
			readFunc: func() (int, error) {
				return 123, nil
			},
			expectPanic: false,
			expectedVal: 123,
		})
	})

	t.Run("Read item panics on error", func(t *testing.T) {
		testErr := "failed to read"
		check(t, testScenario[int]{
			name: "read error causes panic",
			readFunc: func() (int, error) {
				return 0, errors.New(testErr)
			},
			expectPanic: true,
			expectedErr: testErr,
		})
	})
}

func TestUnitMustItemWriter(t *testing.T) {
	type testScenario[T any] struct {
		name        string
		writeFunc   func(item T) error
		inputItem   T
		expectPanic bool
		expectedErr string // Used when expectPanic is true
	}

	check := func(t *testing.T, scenario testScenario[string]) {
		t.Helper()
		mustWriter := NewMustItemReaderWriter(nil, scenario.writeFunc)

		if scenario.expectPanic {
			var recovered any
			func() {
				defer func() {
					recovered = recover()
				}()
				mustWriter.MustWriteItem(scenario.inputItem)
			}()

			require.NotNil(t, recovered, "MustWriteItem should have panicked")
			se, ok := recovered.(errorz.StructuredError)
			require.True(t, ok, "Panic value should be a StructuredError")
			require.Contains(t, se.Unwrap().Error(), scenario.expectedErr, "StructuredError should wrap the expected error")
		} else {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("MustWriteItem should not have panicked, but panicked with: %v", r)
				}
			}()
			mustWriter.MustWriteItem(scenario.inputItem)
		}
	}

	t.Run("Successfully write item", func(t *testing.T) {
		check(t, testScenario[string]{
			name: "successful write",
			writeFunc: func(item string) error {
				require.Equal(t, "test_data", item, "WriteItem should receive the correct item")
				return nil
			},
			inputItem:   "test_data",
			expectPanic: false,
		})
	})

	t.Run("Write item panics on error", func(t *testing.T) {
		testErr := "failed to write"
		check(t, testScenario[string]{
			name: "write error causes panic",
			writeFunc: func(item string) error {
				return errors.New(testErr)
			},
			inputItem:   "some_data",
			expectPanic: true,
			expectedErr: testErr,
		})
	})
}
