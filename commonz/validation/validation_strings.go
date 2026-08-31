package validation

import "strings"

func StrContains(expected string, actual string, msg string, args ...any) error {
	if !strings.Contains(actual, expected) {
		return newError("must contains", map[string]any{"expected": expected, "actual": actual}, msg, args...)
	}
	return nil
}
