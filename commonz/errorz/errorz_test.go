package errorz_test

import (
	"errors"
	"fmt"
	"strings" // Added for string manipulation
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/stretchr/testify/assert"
)

func TestUnitNew(t *testing.T) {
	err := errors.New("test error")
	detailedErr := errorz.New(err)

	assert.NotNil(t, detailedErr)
	assert.Equal(t, "test error", detailedErr.Error())
	assert.NotEmpty(t, detailedErr.StackTrace())
}

func TestUnitNewWithNil(t *testing.T) {
	wrappedErr := errorz.New(nil)
	assert.Nil(t, wrappedErr)
}

func TestUnitUnwrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := errorz.New(originalErr)

	unwrapped := errors.Unwrap(wrappedErr)
	assert.Equal(t, originalErr, unwrapped)
}

// getWrappedErrorFromHelper is a helper function to create a deeper call stack
func getWrappedErrorFromHelper(err error) errorz.DetailedError {
	return errorz.New(err)
}

func TestUnitStackTraceContentAndSkip(t *testing.T) {
	originalErr := errors.New("error from helper")
	detailedErr := getWrappedErrorFromHelper(originalErr)

	assert.NotNil(t, detailedErr)
	stackTrace := detailedErr.StackTrace()
	assert.NotEmpty(t, stackTrace)

	// Assert that the stack trace contains the names of the calling functions
	assert.True(t, strings.Contains(stackTrace, "errorz_test.getWrappedErrorFromHelper"), "Stack trace should contain getWrappedErrorFromHelper")
	assert.True(t, strings.Contains(stackTrace, "errorz_test.TestUnitStackTraceContentAndSkip"), "Stack trace should contain TestUnitStackTraceContentAndSkip")

	// Assert that the stack trace does NOT contain internal errorz.New or runtimez.StackTraceString due to skipping
	assert.False(t, strings.Contains(stackTrace, "errorz.New"), "Stack trace should not contain errorz.New (skipped frame)")
	assert.False(t, strings.Contains(stackTrace, "runtimez.StackTraceString"), "Stack trace should not contain runtimez.StackTraceString (skipped frame)")
	// The detailedError.StackTrace() method itself might appear if not correctly skipped, but New() should skip its own internals.
	// We are checking that New() correctly skips frames that are part of its own creation process.
	assert.False(t, strings.Contains(stackTrace, "errorz.(*detailedError).StackTrace"), "Stack trace should not contain errorz.(*detailedError).StackTrace when capturing the stack for New()")

	// Verify the original error is still unwrappable
	unwrapped := errors.Unwrap(detailedErr)
	assert.Equal(t, originalErr, unwrapped)
}

func TestUnitStackTraceFunction_WithDetailedError(t *testing.T) {
	detailedErr := errorz.New(errors.New("i'm a detailed error"))
	stack := errorz.StackTrace(detailedErr)
	assert.NotEmpty(t, stack)
	assert.Equal(t, detailedErr.StackTrace(), stack)
}

func TestUnitStackTraceFunction_WithWrappedDetailedError(t *testing.T) {
	detailedErr := errorz.New(errors.New("i'm a detailed error"))
	wrappedErr := fmt.Errorf("i'm wrapping a detailed error: %w", detailedErr)

	stack := errorz.StackTrace(wrappedErr)
	assert.NotEmpty(t, stack)
	assert.Equal(t, detailedErr.StackTrace(), stack)
}

func TestUnitStackTraceFunction_WithStandardError(t *testing.T) {
	err := errors.New("i'm a standard error")
	stack := errorz.StackTrace(err)
	assert.Empty(t, stack)
}

func TestUnitStackTraceFunction_WithNilError(t *testing.T) {
	stack := errorz.StackTrace(nil)
	assert.Empty(t, stack)
}

