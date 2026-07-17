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

func TestUnitErrorFormattingAllFields(t *testing.T) {
	baseErr := errors.New("base error")
	err := errorz.Detail(500, "TestName", "payload", true, baseErr)
	expected := "base error: (TestName, code=500, payload=payload, business=true)"
	assert.Equal(t, expected, err.Error())
}

func TestUnitErrorFormattingNoOptionalFields(t *testing.T) {
	baseErr := errors.New("base error")
	err := errorz.Detail(0, "", "", false, baseErr)
	expected := "base error"
	assert.Equal(t, expected, err.Error())
}

func TestUnitErrorFormattingOnlyBusiness(t *testing.T) {
	baseErr := errors.New("base error")
	err := errorz.Detail(0, "", "", true, baseErr)
	expected := "base error: (business=true)"
	assert.Equal(t, expected, err.Error())
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

func TestUnitErrorsAsDirectStructuredError(t *testing.T) {
	se := errorz.Detail(100, "DirectError", "", false, errors.New("base"))
	var target errorz.StructuredError
	assert.True(t, errors.As(se, &target))
	assert.Equal(t, se, target)
}

func TestUnitErrorsAsStructuredErrorWrappedOnce(t *testing.T) {
	se := errorz.Detail(101, "WrappedOnce", "", false, errors.New("base"))
	wrappedErr := fmt.Errorf("layer 1: %w", se)
	var target errorz.StructuredError
	assert.True(t, errors.As(wrappedErr, &target))
	assert.Equal(t, se, target)
}

func TestUnitErrorsAsStructuredErrorWrappedMultipleTimes(t *testing.T) {
	se := errorz.Detail(102, "WrappedMultiple", "", false, errors.New("base"))
	wrappedErr := fmt.Errorf("layer 1: %w", se)
	doubleWrappedErr := fmt.Errorf("layer 2: %w", wrappedErr)
	tripleWrappedErr := fmt.Errorf("layer 3: %w", doubleWrappedErr)

	var target errorz.StructuredError
	assert.True(t, errors.As(tripleWrappedErr, &target))
	assert.Equal(t, se, target)
}

func TestUnitErrorsAsErrorChainWithoutStructuredError(t *testing.T) {
	err := fmt.Errorf("layer 1: %w", errors.New("standard error"))
	var target errorz.StructuredError
	assert.False(t, errors.As(err, &target))
	assert.Nil(t, target)
}

func TestUnitErrorsAsCheckingForDifferentConcreteErrorTypeWithStructuredErrorInChain(t *testing.T) {
	type customError struct{ error }
	baseErr := customError{errors.New("custom base")}
	se := errorz.Detail(103, "WithCustom", "", false, baseErr)

	var targetCustom customError
	assert.True(t, errors.As(se, &targetCustom))
	assert.Equal(t, baseErr, targetCustom)

	var targetStructured errorz.StructuredError
	assert.True(t, errors.As(se, &targetStructured))
	assert.Equal(t, se, targetStructured)
}

func TestUnitErrorsAsWithNilError(t *testing.T) {
	ret := errorz.As(nil)
	assert.Nil(t, ret)
}

func TestUnitAsWithStandardError(t *testing.T) {
	stdErr := errors.New("standard error")
	ret := errorz.As(stdErr)
	assert.NotNil(t, ret)
	assert.Equal(t, "InternalError", ret.Name())
	assert.Equal(t, 500, ret.Code())
	assert.Equal(t, stdErr, errors.Unwrap(ret))
}

func TestUnitAsWithStructuredError(t *testing.T) {
	structuredErr := errorz.Detail(200, "SomeStructuredError", "payload", true, errors.New("base"))
	ret := errorz.As(structuredErr)
	assert.NotNil(t, ret)
	assert.Equal(t, structuredErr, ret)
	assert.Equal(t, 200, ret.Code())
	assert.Equal(t, "SomeStructuredError", ret.Name())
}
