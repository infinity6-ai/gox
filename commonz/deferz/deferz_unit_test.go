package deferz_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/infinity6-ai/gox/commonz/deferz"
)

// mockCloser is a mock implementation of io.Closer for testing purposes.
type mockCloser struct {
	closed bool
	err    error
}

func (m *mockCloser) Close() error {
	m.closed = true
	return m.err
}

func TestUnitNew(t *testing.T) {
	t.Parallel()

	type testScenario struct {
		name string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		d := deferz.New(context.Background())
		require.NotNil(t, d, "New should return a non-nil Deferz instance")
	}

	t.Run("Successfully creates a new Deferz instance", func(t *testing.T) {
		check(t, testScenario{name: "Default creation"})
	})
}

func TestUnitAdd(t *testing.T) {
	t.Parallel()

	type testScenario struct {
		name          string
		numFuncsToAdd int
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		d := deferz.New(context.Background())
		var counter int
		for i := 0; i < s.numFuncsToAdd; i++ {
			d.Add(func() {
				counter++
			})
		}

		// Verify that functions are added but not yet executed
		// The internal state of entries is not directly accessible, so we rely on Do() to verify
		d.Do()
		require.Equal(t, s.numFuncsToAdd, counter, "All added functions should have been executed")
	}

	t.Run("Add a single function", func(t *testing.T) {
		check(t, testScenario{name: "Single function", numFuncsToAdd: 1})
	})

	t.Run("Add multiple functions", func(t *testing.T) {
		check(t, testScenario{name: "Multiple functions", numFuncsToAdd: 5})
	})

	t.Run("Add zero functions", func(t *testing.T) {
		check(t, testScenario{name: "Zero functions", numFuncsToAdd: 0})
	})
}

func TestUnitAddCloser(t *testing.T) {
	t.Parallel()

	type testScenario struct {
		name             string
		closerError      error
		expectResultCall bool
		expectedResult   string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		d := deferz.New(context.Background())
		mock := &mockCloser{err: s.closerError}
		var resultCalled bool
		var receivedErr string

		resultFn := func(err error) {
			resultCalled = true
			if err != nil {
				receivedErr = err.Error()
			}
		}

		if s.expectResultCall {
			d.AddCloser(mock, resultFn)
		} else {
			d.AddCloser(mock, nil)
		}

		require.False(t, mock.closed, "Closer should not be closed before Do()")
		d.Do()
		require.True(t, mock.closed, "Closer should be closed after Do()")

		if s.expectResultCall {
			require.True(t, resultCalled, "Result function should be called")
			if s.closerError != nil {
				require.Equal(t, s.expectedResult, receivedErr, "Result function should receive the closer's error")
			} else {
				require.Empty(t, receivedErr, "Result function should receive no error for successful close")
			}
		} else {
			require.False(t, resultCalled, "Result function should not be called if nil was passed")
		}
	}

	t.Run("AddCloser with successful close and result function", func(t *testing.T) {
		check(t, testScenario{
			name:             "Successful close with result",
			closerError:      nil,
			expectResultCall: true,
		})
	})

	t.Run("AddCloser with error during close and result function", func(t *testing.T) {
		check(t, testScenario{
			name:             "Error close with result",
			closerError:      io.ErrUnexpectedEOF,
			expectResultCall: true,
			expectedResult:   io.ErrUnexpectedEOF.Error(),
		})
	})

	t.Run("AddCloser with successful close and no result function", func(t *testing.T) {
		check(t, testScenario{
			name:             "Successful close without result",
			closerError:      nil,
			expectResultCall: false,
		})
	})
}

func TestUnitAddCloserS(t *testing.T) {
	t.Parallel()

	type testScenario struct {
		name        string
		closerError error
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		d := deferz.New(context.Background())
		mock := &mockCloser{err: s.closerError}

		d.AddCloserS(mock)

		require.False(t, mock.closed, "Closer should not be closed before Do()")
		d.Do()
		require.True(t, mock.closed, "Closer should be closed after Do()")
	}

	t.Run("AddCloserS with successful close", func(t *testing.T) {
		check(t, testScenario{name: "Successful close", closerError: nil})
	})

	t.Run("AddCloserS with error during close", func(t *testing.T) {
		check(t, testScenario{name: "Error close", closerError: io.ErrClosedPipe})
	})
}

func TestUnitClean(t *testing.T) {
	t.Parallel()

	type testScenario struct {
		name          string
		numFuncsToAdd int
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		d := deferz.New(context.Background())
		var counter int
		for i := 0; i < s.numFuncsToAdd; i++ {
			d.Add(func() {
				counter++
			})
		}
		require.GreaterOrEqual(t, s.numFuncsToAdd, 0)
		d.Clean()
		d.Do() // Do should not execute anything
		require.Equal(t, 0, counter, "No functions should be executed after Clean()")
	}

	t.Run("Clean after adding functions", func(t *testing.T) {
		check(t, testScenario{name: "With functions", numFuncsToAdd: 3})
	})

	t.Run("Clean when no functions are added", func(t *testing.T) {
		check(t, testScenario{name: "No functions", numFuncsToAdd: 0})
	})
}

func TestUnitDetach(t *testing.T) {
	t.Parallel()

	type testScenario struct {
		name          string
		numFuncsToAdd int
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		d := deferz.New(context.Background())
		var counter1, counter2 int

		for i := 0; i < s.numFuncsToAdd; i++ {
			d.Add(func() {
				counter1++
			})
		}

		detachedD := d.Detach()

		// Original Deferz should be empty
		d.Do()
		require.Equal(t, 0, counter1, "Original Deferz should not execute functions after Detach()")

		// Detached Deferz should execute functions
		detachedD.Do()
		require.Equal(t, s.numFuncsToAdd, counter1, "Detached Deferz should execute all functions")

		// Add new function to original Deferz and ensure it's independent
		d.Add(func() {
			counter2++
		})
		d.Do()
		require.Equal(t, 1, counter2, "New function added to original Deferz should execute independently")
	}

	t.Run("Detach with functions", func(t *testing.T) {
		check(t, testScenario{name: "With functions", numFuncsToAdd: 3})
	})

	t.Run("Detach with no functions", func(t *testing.T) {
		check(t, testScenario{name: "No functions", numFuncsToAdd: 0})
	})
}

func TestUnitDo(t *testing.T) {
	t.Parallel()

	type testScenario struct {
		name          string
		numFuncsToAdd int
		expectedOrder string // To test LIFO
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		d := deferz.New(context.Background())
		var executionOrder strings.Builder

		for i := 1; i <= s.numFuncsToAdd; i++ {
			val := i // Capture loop variable
			d.Add(func() {
				executionOrder.WriteString(string(rune('0' + val)))
			})
		}

		d.Do()
		require.Equal(t, s.expectedOrder, executionOrder.String(), "Functions should execute in LIFO order")

		// Ensure that Do() clears the entries
		executionOrder.Reset()
		d.Do()
		require.Empty(t, executionOrder.String(), "Do() should clear all entries after execution")
	}

	t.Run("Do with multiple functions (LIFO order)", func(t *testing.T) {
		check(t, testScenario{name: "Multiple functions", numFuncsToAdd: 3, expectedOrder: "321"})
	})

	t.Run("Do with single function", func(t *testing.T) {
		check(t, testScenario{name: "Single function", numFuncsToAdd: 1, expectedOrder: "1"})
	})

	t.Run("Do with no functions", func(t *testing.T) {
		check(t, testScenario{name: "No functions", numFuncsToAdd: 0, expectedOrder: ""})
	})
}

func TestUnitClose(t *testing.T) {
	t.Parallel()

	type testScenario struct {
		name          string
		numFuncsToAdd int
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		d := deferz.New(context.Background())
		var counter int
		for i := 0; i < s.numFuncsToAdd; i++ {
			d.Add(func() {
				counter++
			})
		}

		err := d.Close()
		require.NoError(t, err, "Close should not return an error")
		require.Equal(t, s.numFuncsToAdd, counter, "All added functions should have been executed by Close()")

		// Ensure that Close() clears the entries
		counter = 0
		d.Do() // Should do nothing
		require.Equal(t, 0, counter, "Entries should be cleared after Close()")
	}

	t.Run("Close with functions", func(t *testing.T) {
		check(t, testScenario{name: "With functions", numFuncsToAdd: 3})
	})

	t.Run("Close with no functions", func(t *testing.T) {
		check(t, testScenario{name: "No functions", numFuncsToAdd: 0})
	})
}
