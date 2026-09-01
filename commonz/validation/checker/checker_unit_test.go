package checker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitTrue(t *testing.T) {
	t.Run("should not panic when true", func(t *testing.T) {
		require.NotPanics(t, func() {
			True(true, "should be true")
		})
	})

	t.Run("should panic when false", func(t *testing.T) {
		require.Panics(t, func() {
			True(false, "should be true")
		})
	})
}

func TestUnitFalse(t *testing.T) {
	t.Run("should not panic when false", func(t *testing.T) {
		require.NotPanics(t, func() {
			False(false, "should be false")
		})
	})

	t.Run("should panic when true", func(t *testing.T) {
		require.Panics(t, func() {
			False(true, "should be false")
		})
	})
}

func TestUnitEqual(t *testing.T) {
	t.Run("should not panic when equal", func(t *testing.T) {
		require.NotPanics(t, func() {
			Equal(1, 1, "should be equal")
		})
	})

	t.Run("should panic when not equal", func(t *testing.T) {
		require.Panics(t, func() {
			Equal(1, 2, "should be equal")
		})
	})
}

func TestUnitNotEqual(t *testing.T) {
	t.Run("should not panic when not equal", func(t *testing.T) {
		require.NotPanics(t, func() {
			NotEqual(1, 2, "should not be equal")
		})
	})

	t.Run("should panic when equal", func(t *testing.T) {
		require.Panics(t, func() {
			NotEqual(1, 1, "should not be equal")
		})
	})
}

func TestUnitOneOf(t *testing.T) {
	t.Run("should not panic when one of", func(t *testing.T) {
		require.NotPanics(t, func() {
			OneOf([]int{1, 2, 3}, 2, "should be one of")
		})
	})

	t.Run("should panic when not one of", func(t *testing.T) {
		require.Panics(t, func() {
			OneOf([]int{1, 2, 3}, 4, "should be one of")
		})
	})
}

func TestUnitFail(t *testing.T) {
	t.Run("should always panic", func(t *testing.T) {
		require.Panics(t, func() {
			Fail("this should always fail")
		})
	})
}

func TestUnitNotNil(t *testing.T) {
	t.Run("should not panic when not nil", func(t *testing.T) {
		require.NotPanics(t, func() {
			NotNil(new(int), "should not be nil")
		})
	})

	t.Run("should panic when nil", func(t *testing.T) {
		require.Panics(t, func() {
			NotNil(nil, "should not be nil")
		})
	})
}

func TestUnitGreater(t *testing.T) {
	t.Run("should not panic when greater", func(t *testing.T) {
		require.NotPanics(t, func() {
			Greater(2, 1, "should be greater")
		})
	})

	t.Run("should panic when not greater", func(t *testing.T) {
		require.Panics(t, func() {
			Greater(1, 2, "should be greater")
		})
	})

	t.Run("should panic when equal", func(t *testing.T) {
		require.Panics(t, func() {
			Greater(1, 1, "should be greater")
		})
	})
}

func TestUnitGreaterOrEqual(t *testing.T) {
	t.Run("should not panic when greater", func(t *testing.T) {
		require.NotPanics(t, func() {
			GreaterOrEqual(2, 1, "should be greater or equal")
		})
	})

	t.Run("should not panic when equal", func(t *testing.T) {
		require.NotPanics(t, func() {
			GreaterOrEqual(1, 1, "should be greater or equal")
		})
	})

	t.Run("should panic when less", func(t *testing.T) {
		require.Panics(t, func() {
			GreaterOrEqual(1, 2, "should be greater or equal")
		})
	})
}

func TestUnitLess(t *testing.T) {
	t.Run("should not panic when less", func(t *testing.T) {
		require.NotPanics(t, func() {
			Less(1, 2, "should be less")
		})
	})

	t.Run("should panic when not less", func(t *testing.T) {
		require.Panics(t, func() {
			Less(2, 1, "should be less")
		})
	})

	t.Run("should panic when equal", func(t *testing.T) {
		require.Panics(t, func() {
			Less(1, 1, "should be less")
		})
	})
}

func TestUnitLessOrEqual(t *testing.T) {
	t.Run("should not panic when less", func(t *testing.T) {
		require.NotPanics(t, func() {
			LessOrEqual(1, 2, "should be less or equal")
		})
	})

	t.Run("should not panic when equal", func(t *testing.T) {
		require.NotPanics(t, func() {
			LessOrEqual(1, 1, "should be less or equal")
		})
	})

	t.Run("should panic when greater", func(t *testing.T) {
		require.Panics(t, func() {
			LessOrEqual(2, 1, "should be less or equal")
		})
	})
}

func TestUnitStringRegex(t *testing.T) {
	t.Run("should not panic when regex matches", func(t *testing.T) {
		require.NotPanics(t, func() {
			StringRegex(`^\d+$`, "123", "should match regex")
		})
	})

	t.Run("should panic when regex does not match", func(t *testing.T) {
		require.Panics(t, func() {
			StringRegex(`^\d+$`, "abc", "should match regex")
		})
	})
}

func TestUnitEmpty(t *testing.T) {
	t.Run("should not panic when slice is empty", func(t *testing.T) {
		require.NotPanics(t, func() {
			Empty([]int{}, "should be empty")
		})
	})

	t.Run("should panic when slice is not empty", func(t *testing.T) {
		require.Panics(t, func() {
			Empty([]int{1}, "should be empty")
		})
	})
}

func TestUnitNotEmpty(t *testing.T) {
	t.Run("should not panic when slice is not empty", func(t *testing.T) {
		require.NotPanics(t, func() {
			NotEmpty([]int{1}, "should not be empty")
		})
	})

	t.Run("should panic when slice is empty", func(t *testing.T) {
		require.Panics(t, func() {
			NotEmpty([]int{}, "should not be empty")
		})
	})
}

func TestUnitLen(t *testing.T) {
	t.Run("should not panic when length is correct", func(t *testing.T) {
		require.NotPanics(t, func() {
			Len([]int{1, 2, 3}, 3, "should have correct length")
		})
	})

	t.Run("should panic when length is incorrect", func(t *testing.T) {
		require.Panics(t, func() {
			Len([]int{1, 2}, 3, "should have correct length")
		})
	})
}

func TestUnitStrPrefix(t *testing.T) {
	t.Run("should not panic when prefix is present", func(t *testing.T) {
		require.NotPanics(t, func() {
			StrPrefix("pre", "prefix", "should have prefix")
		})
	})

	t.Run("should panic when prefix is not present", func(t *testing.T) {
		require.Panics(t, func() {
			StrPrefix("pre", "something", "should have prefix")
		})
	})
}

func TestUnitStrEmpty(t *testing.T) {
	t.Run("should not panic when string is empty", func(t *testing.T) {
		require.NotPanics(t, func() {
			StrEmpty("", "should be empty")
		})
	})

	t.Run("should panic when string is not empty", func(t *testing.T) {
		require.Panics(t, func() {
			StrEmpty("not empty", "should be empty")
		})
	})
}

func TestUnitStrNotEmpty(t *testing.T) {
	t.Run("should not panic when string is not empty", func(t *testing.T) {
		require.NotPanics(t, func() {
			StrNotEmpty("not empty", "should not be empty")
		})
	})

	t.Run("should panic when string is empty", func(t *testing.T) {
		require.Panics(t, func() {
			StrNotEmpty("", "should not be empty")
		})
	})
}

func TestUnitStrContains(t *testing.T) {
	t.Run("should not panic when string contains substring", func(t *testing.T) {
		require.NotPanics(t, func() {
			StrContains("sub", "substring", "should contain substring")
		})
	})

	t.Run("should panic when string does not contain substring", func(t *testing.T) {
		require.Panics(t, func() {
			StrContains("sub", "another", "should contain substring")
		})
	})
}

func TestUnitStrNotContains(t *testing.T) {
	t.Run("should not panic when string does not contain substring", func(t *testing.T) {
		require.NotPanics(t, func() {
			StrNotContains("sub", "another", "should not contain substring")
		})
	})

	t.Run("should panic when string contains substring", func(t *testing.T) {
		require.Panics(t, func() {
			StrNotContains("sub", "substring", "should not contain substring")
		})
	})
}
