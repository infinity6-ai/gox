package checker_test

import (
	"regexp"
	"testing"

	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/stretchr/testify/require"
)

func TestUnitTrue(t *testing.T) {
	t.Run("should not panic when true", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.True(true, "should be true")
		})
	})

	t.Run("should panic when false", func(t *testing.T) {
		require.Panics(t, func() {
			checker.True(false, "should be true")
		})
	})
}

func TestUnitFalse(t *testing.T) {
	t.Run("should not panic when false", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.False(false, "should be false")
		})
	})

	t.Run("should panic when true", func(t *testing.T) {
		require.Panics(t, func() {
			checker.False(true, "should be false")
		})
	})
}

func TestUnitEqual(t *testing.T) {
	t.Run("should not panic when equal", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.Equal(1, 1, "should be equal")
		})
	})

	t.Run("should panic when not equal", func(t *testing.T) {
		require.Panics(t, func() {
			checker.Equal(1, 2, "should be equal")
		})
	})
}

func TestUnitNotEqual(t *testing.T) {
	t.Run("should not panic when not equal", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.NotEqual(1, 2, "should not be equal")
		})
	})

	t.Run("should panic when equal", func(t *testing.T) {
		require.Panics(t, func() {
			checker.NotEqual(1, 1, "should not be equal")
		})
	})
}

func TestUnitOneOf(t *testing.T) {
	t.Run("should not panic when value is one of expected", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.OneOf([]int{1, 2, 3}, 2, "should be one of")
		})
	})

	t.Run("should panic when value is not one of expected", func(t *testing.T) {
		require.Panics(t, func() {
			checker.OneOf([]int{1, 2, 3}, 4, "should be one of")
		})
	})
}

func TestUnitFail(t *testing.T) {
	t.Run("should always panic", func(t *testing.T) {
		require.Panics(t, func() {
			checker.Fail("this should fail")
		})
	})
}

func TestUnitNotNil(t *testing.T) {
	t.Run("should not panic when value is not nil", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.NotNil(new(int), "should not be nil")
		})
	})

	t.Run("should panic when value is nil", func(t *testing.T) {
		require.Panics(t, func() {
			checker.NotNil(nil, "should not be nil")
		})
	})
}

func TestUnitGreater(t *testing.T) {
	t.Run("should not panic when value is greater", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.Greater(2, 1, "should be greater")
		})
	})

	t.Run("should panic when value is not greater", func(t *testing.T) {
		require.Panics(t, func() {
			checker.Greater(1, 2, "should be greater")
		})
	})
}

func TestUnitGreaterOrEqual(t *testing.T) {
	t.Run("should not panic when value is greater or equal", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.GreaterOrEqual(2, 1, "should be greater or equal")
			checker.GreaterOrEqual(2, 2, "should be greater or equal")
		})
	})

	t.Run("should panic when value is less", func(t *testing.T) {
		require.Panics(t, func() {
			checker.GreaterOrEqual(1, 2, "should be greater or equal")
		})
	})
}

func TestUnitLess(t *testing.T) {
	t.Run("should not panic when value is less", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.Less(1, 2, "should be less")
		})
	})

	t.Run("should panic when value is not less", func(t *testing.T) {
		require.Panics(t, func() {
			checker.Less(2, 1, "should be less")
		})
	})
}

func TestUnitLessOrEqual(t *testing.T) {
	t.Run("should not panic when value is less or equal", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.LessOrEqual(1, 2, "should be less or equal")
			checker.LessOrEqual(2, 2, "should be less or equal")
		})
	})

	t.Run("should panic when value is greater", func(t *testing.T) {
		require.Panics(t, func() {
			checker.LessOrEqual(2, 1, "should be less or equal")
		})
	})
}

func TestUnitRegexMatch(t *testing.T) {
	pattern := regexp.MustCompile("^[a-z]+$")

	t.Run("should not panic when value matches regex", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.RegexMatch(pattern, "abc", "should match regex")
		})
	})

	t.Run("should panic when value does not match regex", func(t *testing.T) {
		require.Panics(t, func() {
			checker.RegexMatch(pattern, "123", "should match regex")
		})
	})
}

func TestUnitEmpty(t *testing.T) {
	t.Run("should not panic when slice is empty", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.Empty([]int{}, "should be empty")
		})
	})

	t.Run("should panic when slice is not empty", func(t *testing.T) {
		require.Panics(t, func() {
			checker.Empty([]int{1}, "should be empty")
		})
	})
}

func TestUnitNotEmpty(t *testing.T) {
	t.Run("should not panic when slice is not empty", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.NotEmpty([]int{1}, "should not be empty")
		})
	})

	t.Run("should panic when slice is empty", func(t *testing.T) {
		require.Panics(t, func() {
			checker.NotEmpty([]int{}, "should not be empty")
		})
	})
}

func TestUnitLen(t *testing.T) {
	t.Run("should not panic when slice has correct length", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.Len([]int{1, 2, 3}, 3, "should have length 3")
		})
	})

	t.Run("should panic when slice has incorrect length", func(t *testing.T) {
		require.Panics(t, func() {
			checker.Len([]int{1, 2}, 3, "should have length 3")
		})
	})
}

func TestUnitStrPrefix(t *testing.T) {
	t.Run("should not panic when string has prefix", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.StrPrefix("pre", "prefix", "should have prefix")
		})
	})

	t.Run("should panic when string does not have prefix", func(t *testing.T) {
		require.Panics(t, func() {
			checker.StrPrefix("pre", "wrong", "should have prefix")
		})
	})
}

func TestUnitStrEmpty(t *testing.T) {
	t.Run("should not panic when string is empty", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.StrEmpty("", "should be empty")
		})
	})

	t.Run("should panic when string is not empty", func(t *testing.T) {
		require.Panics(t, func() {
			checker.StrEmpty("not empty", "should be empty")
		})
	})
}

func TestUnitStrNotEmpty(t *testing.T) {
	t.Run("should not panic when string is not empty", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.StrNotEmpty("not empty", "should not be empty")
		})
	})

	t.Run("should panic when string is empty", func(t *testing.T) {
		require.Panics(t, func() {
			checker.StrNotEmpty("", "should not be empty")
		})
	})
}

func TestUnitStrContains(t *testing.T) {
	t.Run("should not panic when string contains substring", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.StrContains("sub", "substring", "should contain substring")
		})
	})

	t.Run("should panic when string does not contain substring", func(t *testing.T) {
		require.Panics(t, func() {
			checker.StrContains("sub", "string", "should contain substring")
		})
	})
}

func TestUnitStrNotContains(t *testing.T) {
	t.Run("should not panic when string does not contain substring", func(t *testing.T) {
		require.NotPanics(t, func() {
			checker.StrNotContains("sub", "string", "should not contain substring")
		})
	})

	t.Run("should panic when string contains substring", func(t *testing.T) {
		require.Panics(t, func() {
			checker.StrNotContains("sub", "substring", "should not contain substring")
		})
	})
}
