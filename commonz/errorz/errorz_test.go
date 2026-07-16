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

func TestUnitBusinessFunction(t *testing.T) {
	// 1. Business(nil)
	t.Run("nil error", func(t *testing.T) {
		result := errorz.Business(nil)
		assert.Nil(t, result)
	})

	// 2. Business(standard error)
	t.Run("standard error", func(t *testing.T) {
		stdErr := errors.New("something went wrong")
		result := errorz.Business(stdErr)
		assert.NotNil(t, result)
		assert.False(t, result.Business())
		assert.Equal(t, 500, result.Code())
		assert.Equal(t, "InternalError", result.Name())
		assert.Equal(t, "", result.Payload())
		// The cause of the returned error should be the original standard error
		assert.Equal(t, stdErr, errors.Unwrap(result))
	})

	// 3. Business(StructuredError, Business=true)
	t.Run("business StructuredError", func(t *testing.T) {
		businessErr := errorz.Detail(400, "UserInputInvalid", "invalid ID", true, errors.New("invalid user ID"))
		result := errorz.Business(businessErr)
		assert.NotNil(t, result)
		assert.True(t, result.Business())
		assert.Equal(t, businessErr, result) // Should return the same error
	})

	// 4. Business(StructuredError, Business=false)
	t.Run("non-business StructuredError", func(t *testing.T) {
		nonBusinessErr := errorz.Detail(500, "DBConnection", "conn string", false, errors.New("db connection failed"))
		result := errorz.Business(nonBusinessErr)
		assert.NotNil(t, result)
		assert.False(t, result.Business())
		assert.Equal(t, 500, result.Code())          // Code retained
		assert.Equal(t, "DBConnection", result.Name()) // Name retained
		assert.Equal(t, "", result.Payload())         // Payload cleared
		// Cause should be generic "InternalError"
		assert.Equal(t, errors.New("InternalError").Error(), errors.Unwrap(result).Error())
		// Make sure it's a *new* StructuredError, not the original
		assert.NotEqual(t, nonBusinessErr, result)
	})

	// 5. Business(wrapped standard error) - behaves like (2)
	t.Run("wrapped standard error", func(t *testing.T) {
		wrappedStdErr := fmt.Errorf("outer layer: %w", errors.New("inner problem"))
		result := errorz.Business(wrappedStdErr)
		assert.NotNil(t, result)
		assert.False(t, result.Business())
		assert.Equal(t, 500, result.Code())
		assert.Equal(t, "InternalError", result.Name())
		assert.Equal(t, "", result.Payload())
		// Unwrap twice to get the original cause
		assert.Equal(t, errors.New("inner problem"), errors.Unwrap(errors.Unwrap(result)))
	})

	// 6. Business(wrapped StructuredError, Business=true)
	t.Run("wrapped business StructuredError", func(t *testing.T) {
		businessErr := errorz.Detail(400, "UserInputInvalid", "invalid ID", true, errors.New("invalid user ID"))
		wrappedBusinessErr := fmt.Errorf("validation layer: %w", businessErr)
		result := errorz.Business(wrappedBusinessErr)
		assert.NotNil(t, result)
		assert.True(t, result.Business())
		assert.Equal(t, businessErr, result) // Should return the underlying business error
	})

	// 7. Business(wrapped StructuredError, Business=false)
	t.Run("wrapped non-business StructuredError", func(t *testing.T) {
		nonBusinessErr := errorz.Detail(500, "DBConnection", "conn string", false, errors.New("db connection failed"))
		wrappedNonBusinessErr := fmt.Errorf("service layer: %w", nonBusinessErr)
		result := errorz.Business(wrappedNonBusinessErr)
		assert.NotNil(t, result)
		assert.False(t, result.Business())
		assert.Equal(t, 500, result.Code())          // Code retained
		assert.Equal(t, "DBConnection", result.Name()) // Name retained
		assert.Equal(t, "", result.Payload())         // Payload cleared
		// Cause should be generic "InternalError"
		assert.Equal(t, errors.New("InternalError").Error(), errors.Unwrap(result).Error())
		assert.NotEqual(t, nonBusinessErr, result) // Should be a new error
	})
}
