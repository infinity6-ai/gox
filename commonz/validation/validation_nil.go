package validation

import "reflect"

func isNil(value any) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func NotNil(value any, msg string, args ...any) error {
	if isNil(value) {
		return newError("validation fail NotNil", map[string]any{"actual": value}, msg, args...)
	}

	return nil
}

func Nil(value any, msg string, args ...any) error {
	if isNil(value) {
		return nil
	}

	return newError("validation fail Nil", map[string]any{"actual": value}, msg, args...)
}
