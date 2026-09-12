package validation

import "slices"

func Equal[T comparable](expected T, actual T, msg string, args ...any) error {
	if expected != actual {
		return newError("must be equal", map[string]any{"expected": expected, "actual": actual}, msg, args...)
	}
	return nil
}

func NotEqual[T comparable](expected T, actual T, msg string, args ...any) error {
	if expected == actual {
		return newError("must not be equal", map[string]any{"actual": actual}, msg, args...)
	}
	return nil
}

func OneOf[T comparable](expected []T, actual T, msg string, args ...any) error {
	if !slices.Contains(expected, actual) {
		return newError("must be one of", map[string]any{"expected": expected, "actual": actual}, msg, args...)
	}
	return nil
}

func Zero[T comparable](actual T, msg string, args ...any) error {
	var zero T
	if actual != zero {
		return newError("must be zero", map[string]any{"actual": actual}, msg, args...)
	}
	return nil
}

func NotZero[T comparable](actual T, msg string, args ...any) error {
	var zero T
	if actual == zero {
		return newError("must not be zero", map[string]any{"actual": actual}, msg, args...)
	}
	return nil
}
