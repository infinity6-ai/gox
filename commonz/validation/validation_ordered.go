package validation

import (
	"golang.org/x/exp/constraints"
)

func Greater[T constraints.Ordered](value, threshold T, msg string, args ...any) error {
	if value <= threshold {
		return newError("must be greater than", map[string]any{"threshold": threshold, "actual": value}, msg, args...)
	}
	return nil
}

func GreaterOrEqual[T constraints.Ordered](value, threshold T, msg string, args ...any) error {
	if value < threshold {
		return newError("must be greater or equal than", map[string]any{"threshold": threshold, "actual": value}, msg, args...)
	}
	return nil
}

func Less[T constraints.Ordered](value, threshold T, msg string, args ...any) error {
	if value >= threshold {
		return newError("must be less than", map[string]any{"threshold": threshold, "actual": value}, msg, args...)
	}
	return nil
}

func LessOrEqual[T constraints.Ordered](value, threshold T, msg string, args ...any) error {
	if value > threshold {
		return newError("must be less or equal than", map[string]any{"threshold": threshold, "actual": value}, msg, args...)
	}
	return nil
}
