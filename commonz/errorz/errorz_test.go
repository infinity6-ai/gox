package errorz_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/stretchr/testify/assert"
)

func TestUnitDataMethod(t *testing.T) {
	originalErr := errors.New("original cause")
	structuredErr := errorz.Detail(500, "DBError", "payload", true, originalErr)

	data := structuredErr.Data()
	expectedData := &errorz.Data{
		Code:     500,
		Name:     "DBError",
		Payload:  "payload",
		Business: true,
		Cause:    "original cause",
		Stack:    structuredErr.StackTrace(),
	}
	assert.Equal(t, expectedData, data)
	assert.NotEmpty(t, data.Stack)
}

func TestUnitDetail(t *testing.T) {
	originalErr := errors.New("database connection failed")
	structuredErr := errorz.Detail(500, "DBError", `{"host":"localhost"}`, true, originalErr)

	assert.NotNil(t, structuredErr)
	assert.Equal(t, `database connection failed: (DBError, code=500, payload={"host":"localhost"}, business=true)`, structuredErr.Error())
	assert.NotEmpty(t, structuredErr.StackTrace())
	assert.Equal(t, 500, structuredErr.Code())
	assert.Equal(t, "DBError", structuredErr.Name())
	assert.Equal(t, `{"host":"localhost"}`, structuredErr.Payload())
	assert.True(t, structuredErr.Business())
}

func TestUnitDetailWithNil(t *testing.T) {
	structuredErr := errorz.Detail(500, "DBError", "", false, nil)
	assert.Nil(t, structuredErr)
}

func TestUnitDetailUnwrap(t *testing.T) {
	originalErr := errors.New("original error")
	structuredErr := errorz.Detail(500, "TestError", "", false, originalErr)

	unwrapped := errors.Unwrap(structuredErr)
	assert.Equal(t, originalErr, unwrapped)
}

func TestUnitDetailf(t *testing.T) {
	structuredErr := errorz.Detailf(404, "NotFound", `{"id":123}`, false, "user %d not found", 123)

	assert.NotNil(t, structuredErr)
	assert.Equal(t, `user 123 not found: (NotFound, code=404, payload={"id":123})`, structuredErr.Error())
	assert.NotEmpty(t, structuredErr.StackTrace())
	assert.Equal(t, 404, structuredErr.Code())
	assert.Equal(t, "NotFound", structuredErr.Name())
	assert.Equal(t, `{"id":123}`, structuredErr.Payload())
	assert.False(t, structuredErr.Business())

	cause := errors.Unwrap(structuredErr)
	assert.NotNil(t, cause)
	assert.Equal(t, "user 123 not found", cause.Error())
	assert.Nil(t, errors.Unwrap(cause))
}

func TestUnitErrorFormatting(t *testing.T) {
	baseErr := errors.New("base error")

	t.Run("all fields", func(t *testing.T) {
		err := errorz.Detail(500, "TestName", "payload", true, baseErr)
		expected := "base error: (TestName, code=500, payload=payload, business=true)"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("no optional fields", func(t *testing.T) {
		err := errorz.Detail(0, "", "", false, baseErr)
		expected := "base error"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("only business", func(t *testing.T) {
		err := errorz.Detail(0, "", "", true, baseErr)
		expected := "base error: (business=true)"
		assert.Equal(t, expected, err.Error())
	})
}

// helper function to create a deeper call stack for stack trace testing
func getDetailedErrorFromHelper(err error) errorz.StructuredError {
	return errorz.Detail(500, "HelperError", "", false, err)
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
	detailedErr := errorz.Detail(500, "Test", "", false, errors.New("i'm a structured error"))
	stack := errorz.StackTrace(detailedErr)
	assert.NotEmpty(t, stack)
	assert.Equal(t, detailedErr.StackTrace(), stack)
}

func TestUnitStackTraceFunction_WithWrappedStructuredError(t *testing.T) {
	detailedErr := errorz.Detail(500, "Test", "", true, errors.New("i'm a structured error"))
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

func TestUnitBusinessFunction_NilError(t *testing.T) {
	result := errorz.Business(nil)
	assert.Nil(t, result)
}

func TestUnitBusinessFunction_StandardError(t *testing.T) {
	stdErr := errors.New("something went wrong")
	result := errorz.Business(stdErr)
	assert.NotNil(t, result)
	assert.False(t, result.Business())
	assert.Equal(t, 500, result.Code())
	assert.Equal(t, "InternalError", result.Name())
	assert.Equal(t, "", result.Payload())
	assert.Equal(t, stdErr, errors.Unwrap(result))
}

func TestUnitBusinessFunction_BusinessStructuredError(t *testing.T) {
	businessErr := errorz.Detail(400, "UserInputInvalid", "invalid ID", true, errors.New("invalid user ID"))
	result := errorz.Business(businessErr)
	assert.NotNil(t, result)
	assert.True(t, result.Business())
	assert.Equal(t, businessErr, result)
}

func TestUnitBusinessFunction_NonBusinessStructuredError(t *testing.T) {
	nonBusinessErr := errorz.Detail(500, "DBConnection", "conn string", false, errors.New("db connection failed"))
	result := errorz.Business(nonBusinessErr)
	assert.NotNil(t, result)
	assert.False(t, result.Business())
	assert.Equal(t, 500, result.Code())
	assert.Equal(t, "DBConnection", result.Name())
	assert.Equal(t, "", result.Payload())
	assert.Equal(t, errors.New("InternalError").Error(), errors.Unwrap(result).Error())
	assert.NotEqual(t, nonBusinessErr, result)
}

func TestUnitBusinessFunction_WrappedStandardError(t *testing.T) {
	wrappedStdErr := fmt.Errorf("outer layer: %w", errors.New("inner problem"))
	result := errorz.Business(wrappedStdErr)
	assert.NotNil(t, result)
	assert.False(t, result.Business())
	assert.Equal(t, 500, result.Code())
	assert.Equal(t, "InternalError", result.Name())
	assert.Equal(t, "", result.Payload())
	assert.Equal(t, errors.New("inner problem"), errors.Unwrap(errors.Unwrap(result)))
}

func TestUnitBusinessFunction_WrappedBusinessStructuredError(t *testing.T) {
	businessErr := errorz.Detail(400, "UserInputInvalid", "invalid ID", true, errors.New("invalid user ID"))
	wrappedBusinessErr := fmt.Errorf("validation layer: %w", businessErr)
	result := errorz.Business(wrappedBusinessErr)
	assert.NotNil(t, result)
	assert.True(t, result.Business())
	assert.Equal(t, businessErr, result)
}

func TestUnitBusinessFunction_WrappedNonBusinessStructuredError(t *testing.T) {
	nonBusinessErr := errorz.Detail(500, "DBConnection", "conn string", false, errors.New("db connection failed"))
	wrappedNonBusinessErr := fmt.Errorf("service layer: %w", nonBusinessErr)
	result := errorz.Business(wrappedNonBusinessErr)
	assert.NotNil(t, result)
	assert.False(t, result.Business())
	assert.Equal(t, 500, result.Code())
	assert.Equal(t, "DBConnection", result.Name())
	assert.Equal(t, "", result.Payload())
	assert.Equal(t, errors.New("InternalError").Error(), errors.Unwrap(result).Error())
	assert.NotEqual(t, nonBusinessErr, result)
}
