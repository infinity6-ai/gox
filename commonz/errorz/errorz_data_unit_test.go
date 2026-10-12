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

func TestUnitFrom(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		d := errorz.From(nil)
		assert.NotNil(t, d)
		assert.Equal(t, 500, d.Code)
		assert.Equal(t, "InternalError", d.Name)
	})

	t.Run("string panic value", func(t *testing.T) {
		d := errorz.From("something went wrong")
		assert.NotNil(t, d)
		assert.False(t, d.Business)
		assert.Equal(t, 500, d.Code)
		assert.Equal(t, "InternalError", d.Name)
		assert.Equal(t, "something went wrong", d.Cause)
	})

	t.Run("business structured error", func(t *testing.T) {
		err := errorz.Detail(404, "NotFound", "user-1", true, errors.New("user not found"))
		d := errorz.From(err)
		assert.NotNil(t, d)
		assert.True(t, d.Business)
		assert.Equal(t, 404, d.Code)
		assert.Equal(t, "NotFound", d.Name)
		assert.Equal(t, "user-1", d.Payload)
		assert.Equal(t, "user not found", d.Cause)
	})

	t.Run("Data.Internal sanitizes", func(t *testing.T) {
		err := errorz.Detail(500, "DBError", "secret", false, errors.New("connection failed"))
		d := errorz.From(err)
		internal := d.Internal()
		assert.Equal(t, 500, internal.Code)
		assert.Equal(t, "InternalError", internal.Name)
		assert.Equal(t, "InternalError", internal.Cause)
		assert.False(t, internal.Business)
	})
}
