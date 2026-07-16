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
	assert.Equal(t, `database connection failed: (DBError, code=500, payload={"host":"localhost"})`, structuredErr.Error())
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
	assert.Equal(t, `user 123 not found: (NotFound, code=404, payload={"id":123})`, structuredErr.Error())
	assert.NotEmpty(t, structuredErr.StackTrace())
	assert.Equal(t, 404, structuredErr.Code())
	assert.Equal(t, "NotFound", structuredErr.Name())
	assert.Equal(t, `{"id":123}`, structuredErr.Payload())

	cause := errors.Unwrap(structuredErr)
	assert.NotNil(t, cause)
	assert.Equal(t, "user 123 not found", cause.Error())
	assert.Nil(t, errors.Unwrap(cause), "The root cause created by Detailf should not be wrappable")
}

func TestUnitErrorFormatting_AllFields(t *testing.T) {
	err := errorz.Detail(500, "TestName", "payload", errors.New("base error"))
	expected := "base error: (TestName, code=500, payload=payload)"
	assert.Equal(t, expected, err.Error())
}

func TestUnitErrorFormatting_OnlyCode(t *testing.T) {
	err := errorz.Detail(500, "", "", errors.New("base error"))
	expected := "base error: (code=500)"
	assert.Equal(t, expected, err.Error())
}

func TestUnitErrorFormatting_OnlyName(t *testing.T) {
	err := errorz.Detail(0, "TestName", "", errors.New("base error"))
	expected := "base error: (TestName)"
	assert.Equal(t, expected, err.Error())
}

func TestUnitErrorFormatting_OnlyPayload(t *testing.T) {
	err := errorz.Detail(0, "", "payload", errors.New("base error"))
	expected := "base error: (payload=payload)"
	assert.Equal(t, expected, err.Error())
}

func TestUnitErrorFormatting_NoFields(t *testing.T) {
	err := errorz.Detail(0, "", "", errors.New("base error"))
	expected := "base error"
	assert.Equal(t, expected, err.Error())
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

	assert.True(t, strings.Contains(stackTrace, "errorz_test.getDetailedErrorFromHelper"))
	assert.True(t, strings.Contains(stackTrace, "errorz_test.TestUnitStackTraceContentAndSkip"))
	assert.False(t, strings.Contains(stackTrace, "errorz.Detail"))
	assert.False(t, strings.Contains(stackTrace, "runtimez.StackTraceString"))
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
