package validation

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

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
