package validation

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitValidationOrdered(t *testing.T) {
	t.Run("Greater", func(t *testing.T) {
		require.NoError(t, Greater(2, 1, "2 > 1"))
		err := Greater(1, 2, "1 > 2")
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrValidation))
		require.Contains(t, err.Error(), "must be greater than")
	})

	t.Run("GreaterOrEqual", func(t *testing.T) {
		require.NoError(t, GreaterOrEqual(2, 1, "2 >= 1"))
		require.NoError(t, GreaterOrEqual(2, 2, "2 >= 2"))
		err := GreaterOrEqual(1, 2, "1 >= 2")
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrValidation))
		require.Contains(t, err.Error(), "must be greater or equal than")
	})

	t.Run("Less", func(t *testing.T) {
		require.NoError(t, Less(1, 2, "1 < 2"))
		err := Less(2, 1, "2 < 1")
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrValidation))
		require.Contains(t, err.Error(), "must be less than")
	})

	t.Run("LessOrEqual", func(t *testing.T) {
		require.NoError(t, LessOrEqual(1, 2, "1 <= 2"))
		require.NoError(t, LessOrEqual(2, 2, "2 <= 2"))
		err := LessOrEqual(2, 1, "2 <= 1")
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrValidation))
		require.Contains(t, err.Error(), "must be less or equal than")
	})
}
