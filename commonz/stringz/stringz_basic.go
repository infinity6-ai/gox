// Package stringz provides a collection of utility functions for string manipulation,
// including accent removal and specialized formatting.
package stringz

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// RemoveAccents removes diacritical marks (accents) from a string.
// It uses Unicode normalization (NFD) to decompose characters and then
// removes the non-spacing marks. For example, "résumé" becomes "resume".
func RemoveAccents(s string) string {
	t := norm.NFD.String(s)
	b := strings.Builder{}
	b.Grow(len(s))
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// ShortPretty creates a "pretty" or "slugified" version of a string by processing
// it in parts. It retains alphanumeric characters and replaces specific separators
// with a given divider.
//
// The function iterates through the input `phrase`, building a result string.
// It allows up to `partSize` alphanumeric characters in a continuous block. When a
// separator character ('@', '_', '-', '.', ' ') is encountered, it appends the
// `div` string to the result and resets the alphanumeric character counter. Any
// characters that are not alphanumeric or a recognized separator are dropped.
//
// NOTE: This function's name might not fully convey its behavior, especially the
// truncation of parts longer than `partSize`. A name like `SlugifyWithPartLimit`
// might be more descriptive for future consideration.
//
// For example:
//
//	ShortPretty(5, "-", "long.word and another_one")
//
// would result in "long-word-and-anoth".
func ShortPretty(partSize int, div string, phrase string) string {
	var ret strings.Builder
	count := 0

	for _, c := range phrase {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			if count < partSize {
				ret.WriteRune(c)
				count++
			}
		} else {
			if c == '@' || c == '_' || c == '-' || c == '.' || c == ' ' {
				count = 0
				ret.WriteString(div)
			}
		}
	}
	return ret.String()
}
