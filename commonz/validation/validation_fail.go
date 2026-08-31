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
