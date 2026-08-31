package validation

func True(value bool, msg string, args ...any) error {
	if value != true {
		return newError("validation fail True", map[string]any{"actual": value}, msg, args...)
	}
	return nil
}

func False(value bool, msg string, args ...any) error {
	if value != false {
		return newError("validation fail False", map[string]any{"actual": value}, msg, args...)
	}
	return nil
}
