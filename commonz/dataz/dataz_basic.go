// Package dataz provides generic utility functions for working with data types
// like strings and byte slices.
package dataz

import (
	"github.com/infinity6-ai/gox/commonz/constraintz"
)

// Limited truncates a string or byte slice to a specified maximum length.
// If the length of the input `s` is already less than or equal to `max`,
// the original slice/string is returned unchanged. Otherwise, it returns
// a new slice/string containing the first `max` elements.
func Limited[T constraintz.Data](s T, max int) T {
	if len(s) <= max {
		return s
	}
	return s[0:max]
}
