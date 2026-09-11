package slicez

import "fmt"

// GrowLenBy grows the length of the slice by n.
// It panics if n is negative.
func GrowLenBy[S ~[]E, E any](s S, n int) S {
	if n < 0 {
		panic("n must be non-negative")
	}
	if n == 0 {
		return s
	}
	return SetLen(s, len(s)+n)
}

// GrowLenTo grows the length of the slice to n.
// If the slice length is already n or greater, it's a no-op.
func GrowLenTo[S ~[]E, E any](s S, n int) S {
	if len(s) >= n {
		return s
	}
	return SetLen(s, n)
}

// SetLen sets the length of a slice to n.
// If n is smaller than the slice's current length, the slice is truncated.
// If n is larger than the slice's current capacity, a new underlying array is allocated.
// It panics if n is negative.
func SetLen[S ~[]E, E any](s S, n int) S {
	if n < 0 {
		panic("len must be non-negative")
	}
	if cap(s) >= n {
		return s[:n]
	}
	newS := make(S, n)
	copy(newS, s)
	return newS
}

// Update iterates over a slice and applies the function fn to each element.
// The function fn receives the index and a pointer to the element, allowing in-place modification.
// Iteration stops if fn returns false or an error.
func Update[I any](s []I, fn func(i int, v *I) (bool, error)) error {
	for i := range s {
		keepGoing, err := fn(i, &s[i])
		if err != nil {
			return fmt.Errorf("update failed at index %d: %w", i, err)
		}
		if !keepGoing {
			break
		}
	}
	return nil
}

// MustUpdate is like Update but panics if the update function returns an error.
// The function fn receives the index and a pointer to the element.
// Iteration stops if fn returns false.
func MustUpdate[I any](s []I, fn func(i int, v *I) bool) {
	for i := range s {
		if !fn(i, &s[i]) {
			break
		}
	}
}

// Map transforms a slice of type I to a slice of type O using the mapping function fn.
// The function fn can also filter elements. If fn returns false for an element, it is not included in the result.
// If fn returns an error, processing stops and the error is returned.
func Map[I any, O any](s []I, fn func(i int, v I) (O, bool, error)) ([]O, error) {
	if s == nil {
		return nil, nil
	}
	res := make([]O, 0, len(s))
	for i, v := range s {
		newVal, keep, err := fn(i, v)
		if err != nil {
			return nil, fmt.Errorf("map failed at index %d: %w", i, err)
		}
		if keep {
			res = append(res, newVal)
		}
	}
	return res, nil
}

// MustMap is like Map but panics if the mapping function returns an error.
// The function fn can also filter elements. If fn returns false for an element, it is not included in the result.
func MustMap[I any, O any](s []I, fn func(i int, v I) (O, bool)) []O {
	if s == nil {
		return nil
	}
	res := make([]O, 0, len(s))
	for i, v := range s {
		if newVal, keep := fn(i, v); keep {
			res = append(res, newVal)
		}
	}
	return res
}

// Filter filters a slice in-place based on the predicate function fn.
// It returns a sub-slice of the original slice. The elements that are not kept are zeroed out in the original slice.
// If fn returns an error, processing stops and the error is returned.
func Filter[S ~[]I, I any](s S, fn func(i int, v I) (bool, error)) (S, error) {
	if s == nil {
		return nil, nil
	}
	n := 0
	for i, v := range s {
		keep, err := fn(i, v)
		if err != nil {
			return nil, fmt.Errorf("filter failed at index %d: %w", i, err)
		}
		if keep {
			s[n] = v
			n++
		}
	}
	var zero I
	for i := n; i < len(s); i++ {
		s[i] = zero
	}
	return s[:n], nil
}

// MustFilter is like Filter but panics if the predicate function returns an error.
// It filters the slice in-place.
func MustFilter[S ~[]I, I any](s S, fn func(i int, v I) bool) S {
	if s == nil {
		return nil
	}
	n := 0
	for i, v := range s {
		if fn(i, v) {
			s[n] = v
			n++
		}
	}
	var zero I
	for i := n; i < len(s); i++ {
		s[i] = zero
	}
	return s[:n]
}
