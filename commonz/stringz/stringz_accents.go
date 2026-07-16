package stringz

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

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
