package validation

func NotNil(value any, msg string, args ...any) error {
	if value == nil {
		return newError("validation fail NotNil", map[string]any{"actual": value}, msg, args...)
	}

	return nil
}
