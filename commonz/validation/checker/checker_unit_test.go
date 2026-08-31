package checker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitCheckerComparable(t *testing.T) {
	t.Run("EqualSuccess", func(t *testing.T) {
		require.NotPanics(t, func() { Equal(1, 1, "should be equal") })
	})

	t.Run("EqualFailure", func(t *testing.T) {
		require.Panics(t, func() { Equal(1, 2, "should not be equal") })
	})

	t.Run("NotEqualSuccess", func(t *testing.T) {
		require.NotPanics(t, func() { NotEqual(1, 2, "should not be equal") })
	})

	t.Run("NotEqualFailure", func(t *testing.T) {
		require.Panics(t, func() { NotEqual(1, 1, "should be equal") })
	})
}

func TestUnitCheckerOrdered(t *testing.T) {
	t.Run("GreaterSuccess", func(t *testing.T) {
		require.NotPanics(t, func() { Greater(2, 1, "2 > 1") })
	})

	t.Run("GreaterFailure", func(t *testing.T) {
		require.Panics(t, func() { Greater(1, 2, "1 > 2") })
	})

	t.Run("GreaterOrEqualSuccess", func(t *testing.T) {
		require.NotPanics(t, func() { GreaterOrEqual(2, 1, "2 >= 1") })
		require.NotPanics(t, func() { GreaterOrEqual(2, 2, "2 >= 2") })
	})

	t.Run("GreaterOrEqualFailure", func(t *testing.T) {
		require.Panics(t, func() { GreaterOrEqual(1, 2, "1 >= 2") })
	})

	t.Run("LessSuccess", func(t *testing.T) {
		require.NotPanics(t, func() { Less(1, 2, "1 < 2") })
	})

	t.Run("LessFailure", func(t *testing.T) {
		require.Panics(t, func() { Less(2, 1, "2 < 1") })
	})

	t.Run("LessOrEqualSuccess", func(t *testing.T) {
		require.NotPanics(t, func() { LessOrEqual(1, 2, "1 <= 2") })
		require.NotPanics(t, func() { LessOrEqual(2, 2, "2 <= 2") })
	})

	t.Run("LessOrEqualFailure", func(t *testing.T) {
		require.Panics(t, func() { LessOrEqual(2, 1, "2 <= 1") })
	})
}

func TestUnitCheckerFail(t *testing.T) {
	t.Run("Fail", func(t *testing.T) {
		require.Panics(t, func() { Fail("this is a test failure: %s", "some value") })
	})
}
