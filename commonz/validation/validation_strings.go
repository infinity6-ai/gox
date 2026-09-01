package validation

import "strings"

func StrPrefix(expected string, actual string, msg string, args ...any) error {
	if !strings.HasPrefix(actual, expected) {
		return newError("prefix not found", map[string]any{"actual": actual, "prefix": expected})
	}
	return nil
}

func StrEmpty(actual string, msg string, args ...any) error {
	if actual != "" {
		return newError("must be empty", map[string]any{"actual": actual}, msg, args...)
	}
	return nil
}

func StrNotEmpty(actual string, msg string, args ...any) error {
	if actual == "" {
		return newError("must not be empty", nil, msg, args...)
	}
	return nil
}

func StrContains(expected string, actual string, msg string, args ...any) error {
	if !strings.Contains(actual, expected) {
		return newError("must contains", map[string]any{"expected": expected, "actual": actual}, msg, args...)
	}
	return nil
}

func StrNotContains(expected string, actual string, msg string, args ...any) error {
	if strings.Contains(actual, expected) {
		return newError("must not contain", map[string]any{"expected": expected, "actual": actual}, msg, args...)
	}
	return nil
}
