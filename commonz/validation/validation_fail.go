package validation

import (
	"errors"
	"fmt"
	"slices"
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
		keys := make([]string, 0, len(v.Params))
		for k := range v.Params {
			keys = append(keys, k)
		}
		slices.Sort(keys)

		first := true
		for _, k := range keys {
			if !first {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%s=%v", k, v.Params[k])
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
