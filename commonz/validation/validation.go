package validation

import (
	"errors"
	"fmt"

	"golang.org/x/exp/constraints"
)

var ErrValidation = errors.New("validation error")

type ValidationError struct {
	Name   string
	Params map[string]any
}

func (v *ValidationError) Error() string {
	panic("unimplemented")
}

func newError(name string, params map[string]any, msg string, args ...any) error {
	err := &ValidationError{
		Name:   name,
		Params: params,
	}
	return fmt.Errorf("%w %w: %s", ErrValidation, err, fmt.Sprintf(msg, args...))
}

func Fail(msg string, args ...any) error {
	return newError("fail", nil, msg, args...)
}

func Equal[T comparable](expected T, actual T, msg string, args ...any) error {
	if expected != actual {
		return newError("must be equal", map[string]any{"expected": expected, "actual": actual}, msg, args...)
	}
	return nil
}

func Greater[T constraints.Ordered](value, threshold T, msg string, args ...any) error {
	if value <= threshold {
		return newError("must be greater than", map[string]any{"threshold": threshold, "actual": value}, msg, args...)
	}
}
