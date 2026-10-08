package camelz

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

var ErrUnsupported = errors.New("unsupported")

type Parsed struct {
	parts []string
}

func P(s string) *Parsed {
	ret, err := Parse(s)
	errorz.Check(err)
	return ret
}

func MustParse(s string) *Parsed {
	return P(s)
}

// Parse camel, pascal, snake lower or kebab lower into lower case parts.
// Cornercase: ABC = []string{"a", "b", "c"}
func Parse(s string) (*Parsed, error) {
	if s == "" {
		return nil, fmt.Errorf("%w: empty string", ErrUnsupported)
	}

	hasUnderscore := strings.Contains(s, "_")
	hasHyphen := strings.Contains(s, "-")

	if hasUnderscore && hasHyphen {
		return nil, fmt.Errorf("%w: mixed delimiters: %s", ErrUnsupported, s)
	}

	if hasUnderscore {
		return parseDelimited(s, '_')
	}

	if hasHyphen {
		return parseDelimited(s, '-')
	}

	return parseCamelOrPascal(s)
}

func parseDelimited(s string, sep rune) (*Parsed, error) {
	runes := []rune(s)
	if !unicode.IsLower(runes[0]) {
		return nil, fmt.Errorf("%w: delimited lower string must start with a lowercase letter: %s", ErrUnsupported, s)
	}
	if runes[len(runes)-1] == sep {
		return nil, fmt.Errorf("%w: trailing delimiter: %s", ErrUnsupported, s)
	}

	for _, r := range runes {
		if r == sep {
			continue
		}
		if !unicode.IsLower(r) && !unicode.IsDigit(r) {
			return nil, fmt.Errorf("%w: invalid character in lower delimited string %q: %c", ErrUnsupported, s, r)
		}
	}

	parts := strings.Split(s, string(sep))
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("%w: consecutive delimiters in %q", ErrUnsupported, s)
		}
	}

	return &Parsed{parts: parts}, nil
}

func parseCamelOrPascal(s string) (*Parsed, error) {
	runes := []rune(s)
	if !unicode.IsLetter(runes[0]) {
		return nil, fmt.Errorf("%w: must start with a letter: %s", ErrUnsupported, s)
	}

	for _, r := range runes {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return nil, fmt.Errorf("%w: invalid character in camel/pascal string %q: %c", ErrUnsupported, s, r)
		}
	}

	var parts []string
	var cur strings.Builder
	for _, r := range runes {
		if unicode.IsUpper(r) {
			if cur.Len() > 0 {
				parts = append(parts, strings.ToLower(cur.String()))
				cur.Reset()
			}
		}
		cur.WriteRune(r)
	}
	if cur.Len() > 0 {
		parts = append(parts, strings.ToLower(cur.String()))
	}

	return &Parsed{parts: parts}, nil
}

func (p *Parsed) Parts() []string {
	if p == nil {
		return nil
	}
	ret := make([]string, len(p.parts))
	copy(ret, p.parts)
	return ret
}

func (p *Parsed) String() string {
	return p.QL()
}

// To camel string
func (p *Parsed) C() string {
	if p == nil || len(p.parts) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(p.parts[0])
	for _, part := range p.parts[1:] {
		b.WriteString(capitalize(part))
	}
	return b.String()
}

// To pascal string
func (p *Parsed) P() string {
	if p == nil || len(p.parts) == 0 {
		return ""
	}
	var b strings.Builder
	for _, part := range p.parts {
		b.WriteString(capitalize(part))
	}
	return b.String()
}

// To snake upper string
func (p *Parsed) SU() string {
	if p == nil || len(p.parts) == 0 {
		return ""
	}
	return strings.ToUpper(strings.Join(p.parts, "_"))
}

// To snake lower string
func (p *Parsed) SL() string {
	if p == nil || len(p.parts) == 0 {
		return ""
	}
	return strings.Join(p.parts, "_")
}

// To kebab upper string
func (p *Parsed) QU() string {
	if p == nil || len(p.parts) == 0 {
		return ""
	}
	return strings.ToUpper(strings.Join(p.parts, "-"))
}

// To kebab lower string
func (p *Parsed) QL() string {
	if p == nil || len(p.parts) == 0 {
		return ""
	}
	return strings.Join(p.parts, "-")
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
