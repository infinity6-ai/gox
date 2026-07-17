package errorz_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/stretchr/testify/assert"
)

func TestUnitBusinessDataFunction_NilError(t *testing.T) {
	result := errorz.BusinessData(nil)
	assert.Nil(t, result)
}

func TestUnitBusinessDataFunction_StandardError(t *testing.T) {
	stdErr := errors.New("something went wrong")
	result := errorz.BusinessData(stdErr)
	assert.NotNil(t, result)
	assert.False(t, result.Business)
	assert.Equal(t, 500, result.Code)
	assert.Equal(t, "InternalError", result.Name)
	assert.Equal(t, "", result.Payload)
	assert.Equal(t, stdErr.Error(), result.Cause)
	assert.Empty(t, result.Stack) // Stack is always empty in BusinessData output
}

func TestUnitBusinessDataFunction_BusinessStructuredError(t *testing.T) {
	businessErr := errorz.Detail(400, "UserInputInvalid", "invalid ID", true, errors.New("invalid user ID"))
	result := errorz.BusinessData(businessErr)
	assert.NotNil(t, result)
	assert.True(t, result.Business)
	assert.Equal(t, 400, result.Code)
	assert.Equal(t, "UserInputInvalid", result.Name)
	assert.Equal(t, "invalid ID", result.Payload) // Payload should be retained
	assert.Equal(t, "invalid user ID", result.Cause)
	assert.Empty(t, result.Stack) // Stack is always empty in BusinessData output
}

func TestUnitBusinessDataFunction_NonBusinessStructuredError(t *testing.T) {
	nonBusinessErr := errorz.Detail(500, "DBConnection", "conn string", false, errors.New("db connection failed"))
	result := errorz.BusinessData(nonBusinessErr)
	assert.NotNil(t, result)
	assert.False(t, result.Business)
	assert.Equal(t, 500, result.Code)
	assert.Equal(t, "DBConnection", result.Name)
	assert.Equal(t, "", result.Payload) // Payload should be stripped
	assert.Equal(t, errors.New("InternalError").Error(), result.Cause)
	assert.Empty(t, result.Stack) // Stack is always empty in BusinessData output
}

func TestUnitBusinessDataFunction_WrappedStandardError(t *testing.T) {
	wrappedStdErr := fmt.Errorf("outer layer: %w", errors.New("inner problem"))
	result := errorz.BusinessData(wrappedStdErr)
	assert.NotNil(t, result)
	assert.False(t, result.Business)
	assert.Equal(t, 500, result.Code)
	assert.Equal(t, "InternalError", result.Name)
	assert.Equal(t, "", result.Payload)
	assert.Equal(t, wrappedStdErr.Error(), result.Cause)
	assert.Empty(t, result.Stack) // Stack is always empty in BusinessData output
}

func TestUnitBusinessDataFunction_WrappedBusinessStructuredError(t *testing.T) {
	businessErr := errorz.Detail(400, "UserInputInvalid", "invalid ID", true, errors.New("invalid user ID"))
	wrappedBusinessErr := fmt.Errorf("validation layer: %w", businessErr)
	result := errorz.BusinessData(wrappedBusinessErr)
	assert.NotNil(t, result)
	assert.True(t, result.Business)
	assert.Equal(t, 400, result.Code)
	assert.Equal(t, "UserInputInvalid", result.Name)
	assert.Equal(t, "invalid ID", result.Payload) // Payload should be retained
	assert.Equal(t, "invalid user ID", result.Cause)
	assert.Empty(t, result.Stack) // Stack is always empty in BusinessData output
}

func TestUnitBusinessDataFunction_WrappedNonBusinessStructuredError(t *testing.T) {
	nonBusinessErr := errorz.Detail(500, "DBConnection", "conn string", false, errors.New("db connection failed"))
	wrappedNonBusinessErr := fmt.Errorf("service layer: %w", nonBusinessErr)
	result := errorz.BusinessData(wrappedNonBusinessErr)
	assert.NotNil(t, result)
	assert.False(t, result.Business)
	assert.Equal(t, 500, result.Code)
	assert.Equal(t, "DBConnection", result.Name)
	assert.Equal(t, "", result.Payload) // Payload should be stripped
	assert.Equal(t, errors.New("InternalError").Error(), result.Cause)
	assert.Empty(t, result.Stack) // Stack is always empty in BusinessData output
}
