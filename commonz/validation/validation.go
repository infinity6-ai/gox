package validation

import (
	"errors"
	"fmt"
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

// func Equal[T comparable](expected T, actual T, msg string, args ...any) error {
// 	if expected != actual {
// 		return errorf("expected %v, got %v", []any{expected, actual}, msg, args...)
// 	}
// 	return nil
// }

// func Greater[T constraints.Ordered](value, threshold T, msg string, args ...any) {
// 	if value <= threshold {
// 		errorf("expected %v to be greater than %v", []any{value, threshold}, msg, args...)
// 		// util.Panic("validation fail Greater", map[string]any{"msg": fmt.Sprintf(msg, args...), "threshold": threshold, "actual": value})
// 	}
// }
