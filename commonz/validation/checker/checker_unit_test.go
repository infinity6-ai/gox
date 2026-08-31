package checker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitCheckerComparable(t *testing.T) {
	type testScenario struct {
		name          string
		testFn        func()
		shouldPanic   bool
		expectedValue any
		actualValue   any
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.shouldPanic {
			require.Panics(t, s.testFn)
		} else {
			require.NotPanics(t, s.testFn)
		}
	}

	t.Run("Equal", func(t *testing.T) {
		scenarios := []testScenario{
			{
				name:          "EqualPanics",
				testFn:        func() { Equal(1, 2, "should not be equal") },
				shouldPanic:   true,
				expectedValue: 1,
				actualValue:   2,
			},
			{
				name:          "EqualNotPanics",
				testFn:        func() { Equal(1, 1, "should be equal") },
				shouldPanic:   false,
				expectedValue: 1,
				actualValue:   1,
			},
		}

		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) {
				check(t, s)
			})
		}
	})

	t.Run("NotEqual", func(t *testing.T) {
		scenarios := []testScenario{
			{
				name:          "NotEqualPanics",
				testFn:        func() { NotEqual(1, 1, "should be equal") },
				shouldPanic:   true,
				expectedValue: 1,
				actualValue:   1,
			},
			{
				name:          "NotEqualNotPanics",
				testFn:        func() { NotEqual(1, 2, "should not be equal") },
				shouldPanic:   false,
				expectedValue: 1,
				actualValue:   2,
			},
		}

		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) {
				check(t, s)
			})
		}
	})

	t.Run("OneOf", func(t *testing.T) {
		scenarios := []testScenario{
			{
				name:          "OneOfPanics",
				testFn:        func() { OneOf([]int{1, 2}, 3, "3 should be one of [1,2]") },
				shouldPanic:   true,
				expectedValue: []int{1, 2},
				actualValue:   3,
			},
			{
				name:          "OneOfNotPanics",
				testFn:        func() { OneOf([]int{1, 2}, 1, "1 should be one of [1,2]") },
				shouldPanic:   false,
				expectedValue: []int{1, 2},
				actualValue:   1,
			},
		}

		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) {
				check(t, s)
			})
		}
	})
}

func TestUnitCheckerOrdered(t *testing.T) {
	type testScenario struct {
		name        string
		testFn      func()
		shouldPanic bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.shouldPanic {
			require.Panics(t, s.testFn)
		} else {
			require.NotPanics(t, s.testFn)
		}
	}

	t.Run("Greater", func(t *testing.T) {
		scenarios := []testScenario{
			{name: "GreaterPanics", testFn: func() { Greater(1, 2, "1 > 2") }, shouldPanic: true},
			{name: "GreaterNotPanics", testFn: func() { Greater(2, 1, "2 > 1") }, shouldPanic: false},
		}
		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) { check(t, s) })
		}
	})

	t.Run("GreaterOrEqual", func(t *testing.T) {
		scenarios := []testScenario{
			{name: "GreaterOrEqualPanics", testFn: func() { GreaterOrEqual(1, 2, "1 >= 2") }, shouldPanic: true},
			{name: "GreaterOrEqualNotPanics1", testFn: func() { GreaterOrEqual(2, 1, "2 >= 1") }, shouldPanic: false},
			{name: "GreaterOrEqualNotPanics2", testFn: func() { GreaterOrEqual(2, 2, "2 >= 2") }, shouldPanic: false},
		}
		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) { check(t, s) })
		}
	})

	t.Run("Less", func(t *testing.T) {
		scenarios := []testScenario{
			{name: "LessPanics", testFn: func() { Less(2, 1, "2 < 1") }, shouldPanic: true},
			{name: "LessNotPanics", testFn: func() { Less(1, 2, "1 < 2") }, shouldPanic: false},
		}
		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) { check(t, s) })
		}
	})

	t.Run("LessOrEqual", func(t *testing.T) {
		scenarios := []testScenario{
			{name: "LessOrEqualPanics", testFn: func() { LessOrEqual(2, 1, "2 <= 1") }, shouldPanic: true},
			{name: "LessOrEqualNotPanics1", testFn: func() { LessOrEqual(1, 2, "1 <= 2") }, shouldPanic: false},
			{name: "LessOrEqualNotPanics2", testFn: func() { LessOrEqual(2, 2, "2 <= 2") }, shouldPanic: false},
		}
		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) { check(t, s) })
		}
	})
}

func TestUnitCheckerFail(t *testing.T) {
	require.Panics(t, func() { Fail("this is a test failure: %s", "some value") })
}

func TestUnitCheckerSlices(t *testing.T) {
	type testScenario struct {
		name        string
		testFn      func()
		shouldPanic bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.shouldPanic {
			require.Panics(t, s.testFn)
		} else {
			require.NotPanics(t, s.testFn)
		}
	}

	t.Run("Empty", func(t *testing.T) {
		scenarios := []testScenario{
			{name: "EmptyPanics", testFn: func() { Empty([]int{1}, "should be empty") }, shouldPanic: true},
			{name: "EmptyNotPanics", testFn: func() { Empty([]int{}, "should be empty") }, shouldPanic: false},
		}
		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) { check(t, s) })
		}
	})

	t.Run("NotEmpty", func(t *testing.T) {
		scenarios := []testScenario{
			{name: "NotEmptyPanics", testFn: func() { NotEmpty([]int{}, "should not be empty") }, shouldPanic: true},
			{name: "NotEmptyNotPanics", testFn: func() { NotEmpty([]int{1}, "should not be empty") }, shouldPanic: false},
		}
		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) { check(t, s) })
		}
	})

	t.Run("Len", func(t *testing.T) {
		scenarios := []testScenario{
			{name: "LenPanics", testFn: func() { Len([]int{1, 2}, 3, "length should be 3") }, shouldPanic: true},
			{name: "LenNotPanics", testFn: func() { Len([]int{1, 2}, 2, "length should be 2") }, shouldPanic: false},
		}
		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) { check(t, s) })
		}
	})
}

func TestUnitCheckerStrings(t *testing.T) {
	type testScenario struct {
		name        string
		testFn      func()
		shouldPanic bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.shouldPanic {
			require.Panics(t, s.testFn)
		} else {
			require.NotPanics(t, s.testFn)
		}
	}

	t.Run("StrContains", func(t *testing.T) {
		scenarios := []testScenario{
			{name: "StrContainsPanics", testFn: func() { StrContains("a", "b", "b contains a") }, shouldPanic: true},
			{name: "StrContainsNotPanics", testFn: func() { StrContains("a", "abc", "abc contains a") }, shouldPanic: false},
		}
		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) { check(t, s) })
		}
	})

	t.Run("StrEmpty", func(t *testing.T) {
		scenarios := []testScenario{
			{name: "StrEmptyPanics", testFn: func() { StrEmpty("a", "should be empty") }, shouldPanic: true},
			{name: "StrEmptyNotPanics", testFn: func() { StrEmpty("", "should be empty") }, shouldPanic: false},
		}
		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) { check(t, s) })
		}
	})

	t.Run("StrNotEmpty", func(t *testing.T) {
		scenarios := []testScenario{
			{name: "StrNotEmptyPanics", testFn: func() { StrNotEmpty("", "should not be empty") }, shouldPanic: true},
			{name: "StrNotEmptyNotPanics", testFn: func() { StrNotEmpty("a", "should not be empty") }, shouldPanic: false},
		}
		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) { check(t, s) })
		}
	})

	t.Run("StrNotContains", func(t *testing.T) {
		scenarios := []testScenario{
			{name: "StrNotContainsPanics", testFn: func() { StrNotContains("a", "abc", "abc not contains a") }, shouldPanic: true},
			{name: "StrNotContainsNotPanics", testFn: func() { StrNotContains("a", "b", "b not contains a") }, shouldPanic: false},
		}
		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) { check(t, s) })
		}
	})
}

func TestUnitCheckerBool(t *testing.T) {
	type testScenario struct {
		name        string
		testFn      func()
		shouldPanic bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.shouldPanic {
			require.Panics(t, s.testFn)
		} else {
			require.NotPanics(t, s.testFn)
		}
	}

	t.Run("True", func(t *testing.T) {
		scenarios := []testScenario{
			{name: "TruePanics", testFn: func() { True(false, "should be true") }, shouldPanic: true},
			{name: "TrueNotPanics", testFn: func() { True(true, "should be true") }, shouldPanic: false},
		}
		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) { check(t, s) })
		}
	})

	t.Run("False", func(t *testing.T) {
		scenarios := []testScenario{
			{name: "FalsePanics", testFn: func() { False(true, "should be false") }, shouldPanic: true},
			{name: "FalseNotPanics", testFn: func() { False(false, "should be false") }, shouldPanic: false},
		}
		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) { check(t, s) })
		}
	})
}

func TestUnitCheckerRegex(t *testing.T) {
	type testScenario struct {
		name        string
		testFn      func()
		shouldPanic bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.shouldPanic {
			require.Panics(t, s.testFn)
		} else {
			require.NotPanics(t, s.testFn)
		}
	}

	t.Run("StringRegex", func(t *testing.T) {
		scenarios := []testScenario{
			{name: "StringRegexPanics", testFn: func() { StringRegex(`\d+`, "abc", "should match regex") }, shouldPanic: true},
			{name: "StringRegexNotPanics", testFn: func() { StringRegex(`\d+`, "123", "should match regex") }, shouldPanic: false},
		}
		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) { check(t, s) })
		}
	})
}

func TestUnitCheckerNil(t *testing.T) {
	type testScenario struct {
		name        string
		testFn      func()
		shouldPanic bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.shouldPanic {
			require.Panics(t, s.testFn)
		} else {
			require.NotPanics(t, s.testFn)
		}
	}

	t.Run("NotNil", func(t *testing.T) {
		scenarios := []testScenario{
			{name: "NotNilPanics", testFn: func() { NotNil(nil, "should not be nil") }, shouldPanic: true},
			{name: "NotNilNotPanics", testFn: func() { NotNil(1, "should not be nil") }, shouldPanic: false},
		}
		for _, s := range scenarios {
			t.Run(s.name, func(t *testing.T) { check(t, s) })
		}
	})
}
