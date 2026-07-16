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
}

func TestUnitDataMethod(t *testing.T) {
	originalErr := errors.New("original cause")
	structuredErr := errorz.Detail(500, "DBError", "payload", originalErr)

	data := structuredErr.Data()
	expectedData := &errorz.Data{
		Code:    500,
		Name:    "DBError",
		Payload: "payload",
		Cause:   "original cause",
	}
	assert.Equal(t, expectedData, data)
}

func TestUnitDetail(t *testing.T) {
	originalErr := errors.New("database connection failed")
	structuredErr := errorz.Detail(500, "DBError", `{"host":"localhost"}`, originalErr)

	assert.NotNil(t, structuredErr)
	assertErrorJSON(t, structuredErr, errorz.Data{
		Code:    500,
		Name:    "DBError",
		Payload: `{"host":"localhost"}`,
		Cause:   "database connection failed",
	})
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
	assertErrorJSON(t, structuredErr, errorz.Data{
		Code:    404,
		Name:    "NotFound",
		Payload: `{"id":123}`,
		Cause:   "user 123 not found",
	})
	assert.NotEmpty(t, structuredErr.StackTrace())
	assert.Equal(t, 404, structuredErr.Code())
	assert.Equal(t, "NotFound", structuredErr.Name())
	assert.Equal(t, `{"id":123}`, structuredErr.Payload())

	cause := errors.Unwrap(structuredErr)
	assert.NotNil(t, cause)
	assert.Equal(t, "user 123 not found", cause.Error())
	assert.Nil(t, errors.Unwrap(cause), "The root cause created by Detailf should not be wrappable")
}

func TestUnitErrorJSONFormatting(t *testing.T) {
	baseErr := errors.New("base error")

	t.Run("all fields", func(t *testing.T) {
		err := errorz.Detail(500, "TestName", "payload", baseErr)
		assertErrorJSON(t, err, errorz.Data{Code: 500, Name: "TestName", Payload: "payload", Cause: "base error"})
	})

	t.Run("no optional fields", func(t *testing.T) {
		err := errorz.Detail(0, "", "", baseErr)
		// Check the raw JSON string to ensure omitempty works
		expectedJSON := `{"cause":"base error"}`
		assert.JSONEq(t, expectedJSON, err.Error())
	})

	t.Run("only code", func(t *testing.T) {
		err := errorz.Detail(500, "", "", baseErr)
		expectedJSON := `{"code":500,"cause":"base error"}`
		assert.JSONEq(t, expectedJSON, err.Error())
	})

	t.Run("only name", func(t *testing.T) {
		err := errorz.Detail(0, "TestName", "", baseErr)
		expectedJSON := `{"name":"TestName","cause":"base error"}`
		assert.JSONEq(t, expectedJSON, err.Error())
	})

	t.Run("only payload", func(t *testing.T) {
		err := errorz.Detail(0, "", "payload", baseErr)
		expectedJSON := `{"payload":"payload","cause":"base error"}`
		assert.JSONEq(t, expectedJSON, err.Error())
	})
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
