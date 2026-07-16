package stringz

import (
	"strings"
)

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
