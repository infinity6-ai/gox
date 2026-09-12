package checker_test

import (
	"regexp"
	"testing"

	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/stretchr/testify/require"
)

func TestUnitBool(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		require.NotPanics(t, func() { checker.True(true, "must be true") })
		require.Panics(t, func() { checker.True(false, "must be true") })
	})

	t.Run("False", func(t *testing.T) {
		require.NotPanics(t, func() { checker.False(false, "must be false") })
		require.Panics(t, func() { checker.False(true, "must be false") })
	})
}

func TestUnitComparable(t *testing.T) {
	t.Run("Equal", func(t *testing.T) {
		require.NotPanics(t, func() { checker.Equal(1, 1, "must be equal") })
		require.Panics(t, func() { checker.Equal(1, 2, "must be equal") })
	})

	t.Run("NotEqual", func(t *testing.T) {
		require.NotPanics(t, func() { checker.NotEqual(1, 2, "must be not equal") })
		require.Panics(t, func() { checker.NotEqual(1, 1, "must be not equal") })
	})

	t.Run("OneOf", func(t *testing.T) {
		require.NotPanics(t, func() { checker.OneOf([]int{1, 2, 3}, 2, "must be one of") })
		require.Panics(t, func() { checker.OneOf([]int{1, 2, 3}, 4, "must be one of") })
	})

	t.Run("Zero", func(t *testing.T) {
		require.NotPanics(t, func() { checker.Zero(0, "must be zero") })
		require.Panics(t, func() { checker.Zero(1, "must be zero") })
	})

	t.Run("NotZero", func(t *testing.T) {
		require.NotPanics(t, func() { checker.NotZero(1, "must be not zero") })
		require.Panics(t, func() { checker.NotZero(0, "must be not zero") })
	})
}

func TestUnitFail(t *testing.T) {
	t.Run("Fail", func(t *testing.T) {
		require.Panics(t, func() { checker.Fail("this should fail") })
	})
}

func TestUnitNil(t *testing.T) {
	t.Run("NotNil", func(t *testing.T) {
		require.NotPanics(t, func() { checker.NotNil(new(int), "must not be nil") })
		require.Panics(t, func() { checker.NotNil(nil, "must not be nil") })
	})

	t.Run("Nil", func(t *testing.T) {
		require.NotPanics(t, func() { checker.Nil(nil, "must be nil") })
		require.Panics(t, func() { checker.Nil(new(int), "must be nil") })
	})
}

func TestUnitOrdered(t *testing.T) {
	t.Run("Greater", func(t *testing.T) {
		require.NotPanics(t, func() { checker.Greater(5, 4, "must be greater") })
		require.Panics(t, func() { checker.Greater(4, 5, "must be greater") })
		require.Panics(t, func() { checker.Greater(5, 5, "must be greater") })
	})

	t.Run("GreaterOrEqual", func(t *testing.T) {
		require.NotPanics(t, func() { checker.GreaterOrEqual(5, 4, "must be greater or equal") })
		require.NotPanics(t, func() { checker.GreaterOrEqual(5, 5, "must be greater or equal") })
		require.Panics(t, func() { checker.GreaterOrEqual(4, 5, "must be greater or equal") })
	})

	t.Run("Less", func(t *testing.T) {
		require.NotPanics(t, func() { checker.Less(4, 5, "must be less") })
		require.Panics(t, func() { checker.Less(5, 4, "must be less") })
		require.Panics(t, func() { checker.Less(5, 5, "must be less") })
	})

	t.Run("LessOrEqual", func(t *testing.T) {
		require.NotPanics(t, func() { checker.LessOrEqual(4, 5, "must be less or equal") })
		require.NotPanics(t, func() { checker.LessOrEqual(5, 5, "must be less or equal") })
		require.Panics(t, func() { checker.LessOrEqual(5, 4, "must be less or equal") })
	})
}

func TestUnitRegex(t *testing.T) {
	t.Run("RegexMatch", func(t *testing.T) {
		pattern := regexp.MustCompile(`^a.c$`)
		require.NotPanics(t, func() { checker.RegexMatch(pattern, "abc", "must match regex") })
		require.Panics(t, func() { checker.RegexMatch(pattern, "def", "must match regex") })
	})
}

func TestUnitSlices(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		require.NotPanics(t, func() { checker.Empty([]int{}, "must be empty") })
		require.Panics(t, func() { checker.Empty([]int{1}, "must be empty") })
	})

	t.Run("NotEmpty", func(t *testing.T) {
		require.NotPanics(t, func() { checker.NotEmpty([]int{1}, "must not be empty") })
		require.Panics(t, func() { checker.NotEmpty([]int{}, "must not be empty") })
	})

	t.Run("Len", func(t *testing.T) {
		require.NotPanics(t, func() { checker.Len([]int{1, 2, 3}, 3, "must have length 3") })
		require.Panics(t, func() { checker.Len([]int{1, 2}, 3, "must have length 3") })
	})
}

func TestUnitStrings(t *testing.T) {
	t.Run("StrPrefix", func(t *testing.T) {
		require.NotPanics(t, func() { checker.StrPrefix("pre", "prefix", "must have prefix") })
		require.Panics(t, func() { checker.StrPrefix("abc", "def", "must have prefix") })
	})

	t.Run("StrEmpty", func(t *testing.T) {
		require.NotPanics(t, func() { checker.StrEmpty("", "must be empty") })
		require.Panics(t, func() { checker.StrEmpty("abc", "must be empty") })
	})

	t.Run("StrNotEmpty", func(t *testing.T) {
		require.NotPanics(t, func() { checker.StrNotEmpty("abc", "must not be empty") })
		require.Panics(t, func() { checker.StrNotEmpty("", "must not be empty") })
	})

	t.Run("StrContains", func(t *testing.T) {
		require.NotPanics(t, func() { checker.StrContains("b", "abc", "must contain") })
		require.Panics(t, func() { checker.StrContains("d", "abc", "must contain") })
	})

	t.Run("StrNotContains", func(t *testing.T) {
		require.NotPanics(t, func() { checker.StrNotContains("d", "abc", "must not contain") })
		require.Panics(t, func() { checker.StrNotContains("b", "abc", "must not contain") })
	})
}
