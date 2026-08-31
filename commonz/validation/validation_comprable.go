package validation

func Equal[T comparable](expected T, actual T, msg string, args ...any) error {
	if expected != actual {
		return newError("must be equal", map[string]any{"expected": expected, "actual": actual}, msg, args...)
	}
	return nil
}
