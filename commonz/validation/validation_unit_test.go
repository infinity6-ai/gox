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

func TestUnitValidationComparable(t *testing.T) {
	t.Run("Equal", func(t *testing.T) {
		require.NoError(t, Equal(1, 1, "should be equal"))
		err := Equal(1, 2, "should not be equal")
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrValidation))
		require.Contains(t, err.Error(), "must be equal")
		require.Contains(t, err.Error(), "expected=1")
		require.Contains(t, err.Error(), "actual=2")
	})

	t.Run("NotEqual", func(t *testing.T) {
		require.NoError(t, NotEqual(1, 2, "should not be equal"))
		err := NotEqual(1, 1, "should be equal")
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrValidation))
		require.Contains(t, err.Error(), "must not be equal")
		require.Contains(t, err.Error(), "actual=1")
	})
}

func TestUnitValidationOrdered(t *testing.T) {
	t.Run("Greater", func(t *testing.T) {
		require.NoError(t, Greater(2, 1, "2 > 1"))
		err := Greater(1, 2, "1 > 2")
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be greater than")
	})

	t.Run("GreaterOrEqual", func(t *testing.T) {
		require.NoError(t, GreaterOrEqual(2, 1, "2 >= 1"))
		require.NoError(t, GreaterOrEqual(2, 2, "2 >= 2"))
		err := GreaterOrEqual(1, 2, "1 >= 2")
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be greater or equal than")
	})

	t.Run("Less", func(t *testing.T) {
		require.NoError(t, Less(1, 2, "1 < 2"))
		err := Less(2, 1, "2 < 1")
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be less than")
	})

	t.Run("LessOrEqual", func(t *testing.T) {
		require.NoError(t, LessOrEqual(1, 2, "1 <= 2"))
		require.NoError(t, LessOrEqual(2, 2, "2 <= 2"))
		err := LessOrEqual(2, 1, "2 <= 1")
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be less or equal than")
	})
}

func TestUnitValidationFail(t *testing.T) {
	err := Fail("this is a test failure: %s", "some value")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrValidation))
	require.Contains(t, err.Error(), "fail")
	require.Contains(t, err.Error(), "this is a test failure: some value")
}
