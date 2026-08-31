package validation

import (
	"errors"
	"fmt"
	"strings"
)

var ErrValidation = errors.New("validation error")

type ValidationError struct {
	Name   string
	Params map[string]any
}

func (v *ValidationError) Error() string {
	var sb strings.Builder
	sb.WriteString(v.Name)
	if len(v.Params) > 0 {
		sb.WriteString(" (")
		first := true
		for k, val := range v.Params {
			if !first {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%s=%v", k, val)
			first = false
		}
		sb.WriteString(")")
	}
	return sb.String()
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
