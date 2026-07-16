package errorz_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/stretchr/testify/assert"
)

// assertErrorJSON is a helper to verify the JSON output of the Error() method.
func assertErrorJSON(t *testing.T, err error, expectedData errorz.Data) {
	t.Helper()
	var resultData errorz.Data
	e := json.Unmarshal([]byte(err.Error()), &resultData)
	assert.NoError(t, e, "Error() output should be valid JSON")
	assert.Equal(t, expectedData, resultData)
	assert.Empty(t, resultData.Stack, "Stack field should be empty in Error() JSON output")
}

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
	assertErrorJSON(t, structuredErr, errorz.Data{
		Code:     500,
		Name:     "DBError",
		Payload:  `{"host":"localhost"}`,
		Business: true,
		Cause:    "database connection failed",
		Stack:    "",
	})
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
	// Since Business is false, it should be omitted by omitempty
	assertErrorJSON(t, structuredErr, errorz.Data{
		Code:    404,
		Name:    "NotFound",
		Payload: `{"id":123}`,
		Cause:   "user 123 not found",
		Stack:   "",
	})
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

func TestUnitErrorJSONFormatting(t *testing.T) {
	baseErr := errors.New("base error")

	t.Run("all fields", func(t *testing.T) {
		err := errorz.Detail(500, "TestName", "payload", true, baseErr)
		assertErrorJSON(t, err, errorz.Data{Code: 500, Name: "TestName", Payload: "payload", Business: true, Cause: "base error", Stack: ""})
	})

	t.Run("no optional fields", func(t *testing.T) {
		err := errorz.Detail(0, "", "", false, baseErr)
		expectedJSON := `{"cause":"base error"}`
		assert.JSONEq(t, expectedJSON, err.Error())
	})

	t.Run("only business", func(t *testing.T) {
		err := errorz.Detail(0, "", "", true, baseErr)
		expectedJSON := `{"business":true,"cause":"base error"}`
		assert.JSONEq(t, expectedJSON, err.Error())
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
