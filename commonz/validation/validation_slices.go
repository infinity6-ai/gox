package validation

func Empty[S ~[]E, E any](actual S, msg string, args ...any) error {
	if len(actual) != 0 {
		return newError("must be empty", map[string]any{"actual_length": len(actual)}, msg, args...)
	}
	return nil
}

func NotEmpty[S ~[]E, E any](actual S, msg string, args ...any) error {
	if len(actual) == 0 {
		return newError("must not be empty", nil, msg, args...)
	}
	return nil
}

func Len[S ~[]E, E any](actual S, length int, msg string, args ...any) error {
	if len(actual) != length {
		return newError("failed expected length", map[string]any{"expected_length": length, "actual_length": len(actual)}, msg, args...)
	}
	return nil
}
