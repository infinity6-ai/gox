package validation

import (
	"regexp"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

func StringRegex(pattern, value, msg string, args ...any) error {
	matched, err := regexp.MatchString(pattern, value)
	errorz.Check(err)
	if !matched {
		return newError("validation fail StringRegex", map[string]any{"regex": pattern, "actual": value}, msg, args...)
	}
	return nil
}
