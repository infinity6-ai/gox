package validation

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitValidationError(t *testing.T) {
	t.Run("error with no params", func(t *testing.T) {
		err := &ValidationError{Name: "test_error"}
		require.Equal(t, "test_error", err.Error())
	})

	t.Run("error with params", func(t *testing.T) {
		err := &ValidationError{
			Name: "test_error",
			Params: map[string]any{
				"param1": "value1",
				"param2": 123,
			},
		}
		// NOTE: map iteration order is not guaranteed, so we check for contains
		require.Contains(t, err.Error(), "test_error")
		require.Contains(t, err.Error(), "param1=value1")
		require.Contains(t, err.Error(), "param2=123")
	})
}

func TestUnitValidationFail(t *testing.T) {
	err := Fail("this is a test failure: %s", "some value")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrValidation))
	require.Contains(t, err.Error(), "fail")
	require.Contains(t, err.Error(), "this is a test failure: some value")
}
