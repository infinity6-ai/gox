package slicez

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitSetLen(t *testing.T) {
	type testScenario struct {
		name     string
		initial  func() []int
		newLen   int
		expected []int
		panicMsg string
		check    func(t *testing.T, initial, result []int)
	}

	run := func(t *testing.T, s testScenario) {
		t.Helper()
		initial := s.initial()
		if s.panicMsg != "" {
			require.PanicsWithValue(t, s.panicMsg, func() {
				SetLen(initial, s.newLen)
			})
			return
		}

		result := SetLen(initial, s.newLen)

		require.Equal(t, s.expected, result)
		require.Len(t, result, s.newLen)
		if s.check != nil {
			s.check(t, initial, result)
		}
	}

	t.Run("shrink", func(t *testing.T) {
		run(t, testScenario{
			initial:  func() []int { return []int{1, 2, 3, 4, 5} },
			newLen:   3,
			expected: []int{1, 2, 3},
			check: func(t *testing.T, initial, result []int) {
				if len(result) > 0 {
					require.Same(t, &initial[0], &result[0], "should be same underlying array")
				}
			},
		})
	})

	t.Run("grow within capacity", func(t *testing.T) {
		run(t, testScenario{
			initial: func() []int {
				s := make([]int, 3, 10)
				s[0], s[1], s[2] = 1, 2, 3
				return s
			},
			newLen:   7,
			expected: []int{1, 2, 3, 0, 0, 0, 0},
			check: func(t *testing.T, initial, result []int) {
				if len(result) > 0 {
					require.Same(t, &initial[0], &result[0], "should be same underlying array")
				}
			},
		})
	})

	t.Run("grow beyond capacity", func(t *testing.T) {
		run(t, testScenario{
			initial:  func() []int { return []int{1, 2, 3} },
			newLen:   5,
			expected: []int{1, 2, 3, 0, 0},
			check: func(t *testing.T, initial, result []int) {
				if len(initial) > 0 && len(result) > 0 {
					require.NotSame(t, &initial[0], &result[0], "should have allocated a new underlying array")
				}
			},
		})
	})

	t.Run("to zero", func(t *testing.T) {
		run(t, testScenario{
			initial:  func() []int { return []int{1, 2, 3} },
			newLen:   0,
			expected: []int{},
		})
	})

	t.Run("from nil", func(t *testing.T) {
		run(t, testScenario{
			initial:  func() []int { return nil },
			newLen:   3,
			expected: []int{0, 0, 0},
		})
	})

	t.Run("from empty", func(t *testing.T) {
		run(t, testScenario{
			initial:  func() []int { return []int{} },
			newLen:   2,
			expected: []int{0, 0},
		})
	})

	t.Run("negative len", func(t *testing.T) {
		run(t, testScenario{
			initial:  func() []int { return []int{1, 2, 3} },
			newLen:   -1,
			panicMsg: "len must be non-negative",
		})
	})
}

func TestUnitGrowLenTo(t *testing.T) {
	t.Run("grow", func(t *testing.T) {
		s := []int{1, 2, 3}
		result := GrowLenTo(s, 5)
		require.Equal(t, []int{1, 2, 3, 0, 0}, result)
	})
	t.Run("no grow", func(t *testing.T) {
		s := []int{1, 2, 3, 4, 5}
		result := GrowLenTo(s, 3)
		require.Equal(t, []int{1, 2, 3, 4, 5}, result)
		require.Same(t, &s[0], &result[0])
	})
}

func TestUnitGrowLenBy(t *testing.T) {
	t.Run("grow", func(t *testing.T) {
		s := []int{1, 2, 3}
		result := GrowLenBy(s, 2)
		require.Equal(t, []int{1, 2, 3, 0, 0}, result)
	})
	t.Run("no grow", func(t *testing.T) {
		s := []int{1, 2, 3}
		result := GrowLenBy(s, 0)
		require.Equal(t, []int{1, 2, 3}, result)
		require.Same(t, &s[0], &result[0])
	})
	t.Run("panic negative", func(t *testing.T) {
		require.PanicsWithValue(t, "n must be non-negative", func() {
			GrowLenBy([]int{1, 2, 3}, -1)
		})
	})
}

func TestUnitUpdate(t *testing.T) {
	t.Run("update and keep all elements", func(t *testing.T) {
		s := []int{1, 2, 3}
		originalS := s
		res, err := Update(s, func(i int, v *int) (bool, error) {
			*v = *v * 2
			return true, nil
		})
		require.NoError(t, err)
		require.Equal(t, []int{2, 4, 6}, res)
		require.Same(t, &originalS[0], &res[0])
		require.Equal(t, []int{2, 4, 6}, s)
	})

	t.Run("update and remove some elements", func(t *testing.T) {
		s := []int{1, 2, 3, 4}
		res, err := Update(s, func(i int, v *int) (bool, error) {
			*v = *v + 1
			return *v < 4, nil // remove 3+1=4 and 4+1=5
		})
		require.NoError(t, err)
		require.Equal(t, []int{2, 3}, res)
		require.Equal(t, []int{2, 3, 0, 0}, s)
	})

	t.Run("error occurs", func(t *testing.T) {
		s := []int{1, 2, 3}
		expectedErr := errors.New("some error")
		_, err := Update(s, func(i int, v *int) (bool, error) {
			if i == 1 {
				return false, expectedErr
			}
			*v = *v * 2
			return true, nil
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "update failed at index 1")
		require.ErrorIs(t, err, expectedErr)
		require.Equal(t, []int{2, 2, 3}, s) // only first element was updated
	})
}

func TestUnitMustUpdate(t *testing.T) {
	t.Run("update and keep all", func(t *testing.T) {
		s := []int{1, 2, 3}
		res := MustUpdate(s, func(i int, v *int) bool {
			*v = *v * 2
			return true
		})
		require.Equal(t, []int{2, 4, 6}, res)
		require.Equal(t, []int{2, 4, 6}, s)
	})

	t.Run("update and remove some", func(t *testing.T) {
		s := []int{1, 2, 3, 4}
		res := MustUpdate(s, func(i int, v *int) bool {
			*v = *v * 2
			return *v < 5 // remove 3*2=6 and 4*2=8
		})
		require.Equal(t, []int{2, 4}, res)
		require.Equal(t, []int{2, 4, 0, 0}, s)
	})
}

func TestUnitMap(t *testing.T) {
	t.Run("map and filter", func(t *testing.T) {
		s := []int{1, 2, 3, 4}
		// to string, but only for even numbers
		result, err := Map(s, func(i int, v int) (string, bool, error) {
			isEven := v%2 == 0
			return strconv.Itoa(v), isEven, nil
		})
		require.NoError(t, err)
		require.Equal(t, []string{"2", "4"}, result)
	})

	t.Run("error occurs", func(t *testing.T) {
		s := []int{1, 2, 3}
		expectedErr := errors.New("some error")
		_, err := Map(s, func(i int, v int) (string, bool, error) {
			if i == 1 {
				return "", false, expectedErr
			}
			return strconv.Itoa(v), true, nil
		})
		require.Error(t, err)
		require.ErrorIs(t, err, expectedErr)
		require.Contains(t, err.Error(), "map failed at index 1")
	})

	t.Run("nil slice", func(t *testing.T) {
		result, err := Map(nil, func(i int, v int) (string, bool, error) {
			return "", false, nil
		})
		require.NoError(t, err)
		require.Nil(t, result)
	})
}

func TestUnitMustMap(t *testing.T) {
	t.Run("map and filter", func(t *testing.T) {
		s := []int{1, 2, 3, 4}
		// to string, but only for even numbers
		result := MustMap(s, func(i int, v int) (string, bool) {
			isEven := v%2 == 0
			return strconv.Itoa(v), isEven
		})
		require.Equal(t, []string{"2", "4"}, result)
	})

	t.Run("nil slice", func(t *testing.T) {
		result := MustMap(nil, func(i int, v int) (string, bool) {
			return "", false
		})
		require.Nil(t, result)
	})
}

func TestUnitFilter(t *testing.T) {
	t.Run("filter even numbers", func(t *testing.T) {
		s := []int{1, 2, 3, 4, 5, 6}
		sCopy := append([]int(nil), s...)

		result, err := Filter(s, func(i int, v int) (bool, error) {
			return v%2 == 0, nil
		})

		require.NoError(t, err)
		require.Equal(t, []int{2, 4, 6}, result)
		require.Len(t, result, 3)

		if len(result) > 0 {
			require.Same(t, &s[0], &result[0], "should be sub-slice of original")
		}

		// Check that the tail is zeroed
		require.Equal(t, []int{2, 4, 6, 0, 0, 0}, s, "original slice should be modified")
		_ = sCopy
	})

	t.Run("filter with pointers", func(t *testing.T) {
		a, b, c := 1, 2, 3
		s := []*int{&a, &b, &c}

		result, err := Filter(s, func(i int, v *int) (bool, error) {
			return *v > 1, nil
		})
		require.NoError(t, err)
		require.Equal(t, []*int{&b, &c}, result)

		// Check tail is zeroed (nil for pointers)
		require.Equal(t, []*int{&b, &c, nil}, s)
	})

	t.Run("error occurs", func(t *testing.T) {
		s := []int{1, 2, 3}
		expectedErr := errors.New("some error")
		_, err := Filter(s, func(i int, v int) (bool, error) {
			if i == 1 {
				return false, expectedErr
			}
			return true, nil
		})
		require.Error(t, err)
		require.ErrorIs(t, err, expectedErr)
		require.Contains(t, err.Error(), "filter failed at index 1")
	})

	t.Run("nil slice", func(t *testing.T) {
		result, err := Filter(([]int)(nil), func(i int, v int) (bool, error) {
			return true, nil
		})
		require.NoError(t, err)
		require.Nil(t, result)
	})
}

func TestUnitMustFilter(t *testing.T) {
	t.Run("filter even", func(t *testing.T) {
		s := []int{1, 2, 3, 4, 5, 6}
		result := MustFilter(s, func(i int, v int) bool {
			return v%2 == 0
		})
		require.Equal(t, []int{2, 4, 6}, result)
		require.Equal(t, []int{2, 4, 6, 0, 0, 0}, s)
	})

	t.Run("nil slice", func(t *testing.T) {
		result := MustFilter(([]int)(nil), func(i int, v int) bool {
			return true
		})
		require.Nil(t, result)
	})
}

func TestUnitUnique(t *testing.T) {
	type testScenario struct {
		name          string
		initial       []int
		expected      []int
		expectedArray []int
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		var input []int
		if s.initial != nil {
			input = make([]int, len(s.initial))
			copy(input, s.initial)
		}
		cmpInt := func(a, b int) int {
			if a < b {
				return -1
			}
			if a > b {
				return 1
			}
			return 0
		}
		result := Unique(input, cmpInt)
		if s.expected == nil {
			require.Nil(t, result)
			return
		}
		require.Equal(t, s.expected, result)
		if s.expectedArray != nil {
			require.Equal(t, s.expectedArray, input)
		}
		if len(result) > 0 {
			require.Same(t, &input[0], &result[0], "should be sub-slice of original slice")
		}
	}

	t.Run("nil slice", func(t *testing.T) {
		check(t, testScenario{
			initial:  nil,
			expected: nil,
		})
	})

	t.Run("empty slice", func(t *testing.T) {
		check(t, testScenario{
			initial:       []int{},
			expected:      []int{},
			expectedArray: []int{},
		})
	})

	t.Run("single element", func(t *testing.T) {
		check(t, testScenario{
			initial:       []int{42},
			expected:      []int{42},
			expectedArray: []int{42},
		})
	})

	t.Run("already unique elements", func(t *testing.T) {
		check(t, testScenario{
			initial:       []int{1, 2, 3, 4, 5},
			expected:      []int{1, 2, 3, 4, 5},
			expectedArray: []int{1, 2, 3, 4, 5},
		})
	})

	t.Run("consecutive duplicates", func(t *testing.T) {
		check(t, testScenario{
			initial:       []int{1, 1, 2, 2, 3, 3},
			expected:      []int{1, 2, 3},
			expectedArray: []int{1, 2, 3, 0, 0, 0},
		})
	})

	t.Run("global non-consecutive duplicates", func(t *testing.T) {
		check(t, testScenario{
			initial:       []int{1, 2, 3, 1, 2, 4, 3, 5},
			expected:      []int{1, 2, 3, 4, 5},
			expectedArray: []int{1, 2, 3, 4, 5, 0, 0, 0},
		})
	})

	t.Run("all duplicate elements", func(t *testing.T) {
		check(t, testScenario{
			initial:       []int{7, 7, 7, 7},
			expected:      []int{7},
			expectedArray: []int{7, 0, 0, 0},
		})
	})

	t.Run("custom non-comparable struct with comparator", func(t *testing.T) {
		type customItem struct {
			id   int
			tags []string
		}
		items := []customItem{
			{id: 1, tags: []string{"a"}},
			{id: 2, tags: []string{"b"}},
			{id: 1, tags: []string{"c"}},
			{id: 3, tags: []string{"d"}},
		}
		res := Unique(items, func(a, b customItem) int {
			if a.id < b.id {
				return -1
			}
			if a.id > b.id {
				return 1
			}
			return 0
		})
		require.Equal(t, []customItem{
			{id: 1, tags: []string{"a"}},
			{id: 2, tags: []string{"b"}},
			{id: 3, tags: []string{"d"}},
		}, res)
		require.Equal(t, customItem{}, items[3])
	})

	t.Run("preserves custom slice type", func(t *testing.T) {
		type IntSlice []int
		s := IntSlice{1, 2, 1, 3}
		res := Unique(s, func(a, b int) int {
			return a - b
		})
		require.Equal(t, IntSlice{1, 2, 3}, res)
	})
}

func TestUnitUniqueC(t *testing.T) {
	type testScenario struct {
		name          string
		initial       []string
		expected      []string
		expectedArray []string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		var input []string
		if s.initial != nil {
			input = make([]string, len(s.initial))
			copy(input, s.initial)
		}
		result := UniqueC(input)
		if s.expected == nil {
			require.Nil(t, result)
			return
		}
		require.Equal(t, s.expected, result)
		if s.expectedArray != nil {
			require.Equal(t, s.expectedArray, input)
		}
		if len(result) > 0 {
			require.Same(t, &input[0], &result[0], "should be sub-slice of original slice")
		}
	}

	t.Run("nil slice", func(t *testing.T) {
		check(t, testScenario{
			initial:  nil,
			expected: nil,
		})
	})

	t.Run("empty slice", func(t *testing.T) {
		check(t, testScenario{
			initial:       []string{},
			expected:      []string{},
			expectedArray: []string{},
		})
	})

	t.Run("single element", func(t *testing.T) {
		check(t, testScenario{
			initial:       []string{"a"},
			expected:      []string{"a"},
			expectedArray: []string{"a"},
		})
	})

	t.Run("already unique elements", func(t *testing.T) {
		check(t, testScenario{
			initial:       []string{"a", "b", "c"},
			expected:      []string{"a", "b", "c"},
			expectedArray: []string{"a", "b", "c"},
		})
	})

	t.Run("consecutive duplicates", func(t *testing.T) {
		check(t, testScenario{
			initial:       []string{"a", "a", "b", "b", "c"},
			expected:      []string{"a", "b", "c"},
			expectedArray: []string{"a", "b", "c", "", ""},
		})
	})

	t.Run("global non-consecutive duplicates", func(t *testing.T) {
		check(t, testScenario{
			initial:       []string{"a", "b", "a", "c", "b", "d", "a"},
			expected:      []string{"a", "b", "c", "d"},
			expectedArray: []string{"a", "b", "c", "d", "", "", ""},
		})
	})

	t.Run("all duplicate elements", func(t *testing.T) {
		check(t, testScenario{
			initial:       []string{"x", "x", "x"},
			expected:      []string{"x"},
			expectedArray: []string{"x", "", ""},
		})
	})

	t.Run("pointers zeroed out as nil", func(t *testing.T) {
		v1, v2, v3 := 1, 2, 3
		items := []*int{&v1, &v2, &v1, &v3, &v2}
		res := UniqueC(items)
		require.Equal(t, []*int{&v1, &v2, &v3}, res)
		require.Nil(t, items[3])
		require.Nil(t, items[4])
	})

	t.Run("preserves custom slice type", func(t *testing.T) {
		type StringSlice []string
		s := StringSlice{"a", "b", "a"}
		res := UniqueC(s)
		require.Equal(t, StringSlice{"a", "b"}, res)
	})
}
