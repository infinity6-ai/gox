package validation

func Empty[T string | ~[]E, E any | ~map[K]V, K comparable, V any](actual T, msg string, args ...any) error {
	if len(actual) != 0 {
		return newError("must be empty", map[string]any{"actual_length": len(actual)}, msg, args...)
	}
	return nil
}

func NotEmpty[T string | ~[]E, E any | ~map[K]V, K comparable, V any](actual T, msg string, args ...any) error {
	if len(actual) != 0 {
		return newError("must be empty", map[string]any{"actual_length": len(actual)}, msg, args...)
	}
	return nil
}
