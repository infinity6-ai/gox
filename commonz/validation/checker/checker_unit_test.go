package checker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitCheckerComparable(t *testing.T) {
	t.Run("Equal", func(t *testing.T) {
		require.NotPanics(t, func() { Equal(1, 1, "should be equal") })
		require.Panics(t, func() { Equal(1, 2, "should not be equal") })
	})

	t.Run("NotEqual", func(t *testing.T) {
		require.NotPanics(t, func() { NotEqual(1, 2, "should not be equal") })
		require.Panics(t, func() { NotEqual(1, 1, "should be equal") })
	})
}

func TestUnitCheckerOrdered(t *testing.T) {
	t.Run("Greater", func(t *testing.T) {
		require.NotPanics(t, func() { Greater(2, 1, "2 > 1") })
		require.Panics(t, func() { Greater(1, 2, "1 > 2") })
	})

	t.Run("GreaterOrEqual", func(t *testing.T) {
		require.NotPanics(t, func() { GreaterOrEqual(2, 1, "2 >= 1") })
		require.NotPanics(t, func() { GreaterOrEqual(2, 2, "2 >= 2") })
		require.Panics(t, func() { GreaterOrEqual(1, 2, "1 >= 2") })
	})

	t.Run("Less", func(t *testing.T) {
		require.NotPanics(t, func() { Less(1, 2, "1 < 2") })
		require.Panics(t, func() { Less(2, 1, "2 < 1") })
	})

	t.Run("LessOrEqual", func(t *testing.T) {
		require.NotPanics(t, func() { LessOrEqual(1, 2, "1 <= 2") })
		require.NotPanics(t, func() { LessOrEqual(2, 2, "2 <= 2") })
		require.Panics(t, func() { LessOrEqual(2, 1, "2 <= 1") })
	})
}

func TestUnitCheckerFail(t *testing.T) {
	require.Panics(t, func() { Fail("this is a test failure: %s", "some value") })
}
