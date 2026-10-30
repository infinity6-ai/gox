package urlz

import (
	"regexp"
)

var urlSchemeRegex = regexp.MustCompile(`^[a-z0-9]+:`)

// Is checks if the given string represents a URL by matching the scheme pattern ^[a-z0-9]+:.
func Is(u string) bool {
	return urlSchemeRegex.MatchString(u)
}
