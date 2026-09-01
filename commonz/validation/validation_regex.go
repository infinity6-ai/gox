package validation

import (
	"regexp"
)

func RegexMatch(pattern *regexp.Regexp, value, msg string, args ...any) error {
	// matched, err := regexp.MatchString(pattern, value)
	// errorz.Check(err)
	matched := pattern.MatchString(value)
	if !matched {
		return newError("validation fail StringRegex", map[string]any{"regex": pattern, "actual": value}, msg, args...)
	}
	return nil
}
