package errorz_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/stretchr/testify/assert"
)

func TestUnitDetail(t *testing.T) {
	originalErr := errors.New("database connection failed")
	structuredErr := errorz.Detail(500, "DBError", `{"host":"localhost"}`, originalErr)

	assert.NotNil(t, structuredErr)
	assert.Equal(t, "database connection failed", structuredErr.Error())
	assert.NotEmpty(t, structuredErr.StackTrace())
	assert.Equal(t, 500, structuredErr.Code())
	assert.Equal(t, "DBError", structuredErr.Name())
	assert.Equal(t, `{"host":"localhost"}`, structuredErr.Payload())
}

func TestUnitDetailWithNil(t *testing.T) {
	structuredErr := errorz.Detail(500, "DBError", "", nil)
	assert.Nil(t, structuredErr)
}

func TestUnitDetailUnwrap(t *testing.T) {
	originalErr := errors.New("original error")
	structuredErr := errorz.Detail(500, "TestError", "", originalErr)

	unwrapped := errors.Unwrap(structuredErr)
	assert.Equal(t, originalErr, unwrapped)
}

func TestUnitDetailf(t *testing.T) {
	structuredErr := errorz.Detailf(404, "NotFound", `{"id":123}`, "user %d not found", 123)

	assert.NotNil(t, structuredErr)
	assert.Equal(t, "user 123 not found", structuredErr.Error())
	assert.NotEmpty(t, structuredErr.StackTrace())
	assert.Equal(t, 404, structuredErr.Code())
	assert.Equal(t, "NotFound", structuredErr.Name())
	assert.Equal(t, `{"id":123}`, structuredErr.Payload())

	// Unwrap the structured error to get the cause from fmt.Errorf
	cause := errors.Unwrap(structuredErr)
	assert.NotNil(t, cause)
	assert.Equal(t, "user 123 not found", cause.Error())

	// The cause from fmt.Errorf does not wrap another error, so unwrapping it should be nil
	assert.Nil(t, errors.Unwrap(cause), "The root cause created by Detailf should not be wrappable")
}

// helper function to create a deeper call stack for stack trace testing
func getDetailedErrorFromHelper(err error) errorz.StructuredError {
	return errorz.Detail(500, "HelperError", "", err)
}

func TestUnitStackTraceContentAndSkip(t *testing.T) {
	originalErr := errors.New("error from helper")
	detailedErr := getDetailedErrorFromHelper(originalErr)

	stackTrace := detailedErr.StackTrace()
	assert.NotEmpty(t, stackTrace)

	// Assert that the stack trace contains the names of the calling functions
	assert.True(t, strings.Contains(stackTrace, "errorz_test.getDetailedErrorFromHelper"), "Stack trace should contain getDetailedErrorFromHelper")
	assert.True(t, strings.Contains(stackTrace, "errorz_test.TestUnitStackTraceContentAndSkip"), "Stack trace should contain TestUnitStackTraceContentAndSkip")

	// Assert that the stack trace does NOT contain internal functions due to skipping
	assert.False(t, strings.Contains(stackTrace, "errorz.Detail"), "Stack trace should not contain errorz.Detail")
	assert.False(t, strings.Contains(stackTrace, "runtimez.StackTraceString"), "Stack trace should not contain runtimez.StackTraceString")
}

func TestUnitStackTraceFunction_WithStructuredError(t *testing.T) {
	detailedErr := errorz.Detail(500, "Test", "", errors.New("i'm a structured error"))
	stack := errorz.StackTrace(detailedErr)
	assert.NotEmpty(t, stack)
	assert.Equal(t, detailedErr.StackTrace(), stack)
}

func TestUnitStackTraceFunction_WithWrappedStructuredError(t *testing.T) {
	detailedErr := errorz.Detail(500, "Test", "", errors.New("i'm a structured error"))
	wrappedErr := fmt.Errorf("i'm wrapping a structured error: %w", detailedErr)

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
