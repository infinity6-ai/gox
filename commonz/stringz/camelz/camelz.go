// Package camelz provides case conversion and parsing utilities for strings across
// camelCase, PascalCase, snake_case, and kebab-case conventions.
package camelz

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

// ErrUnsupported indicates that an input string is empty, contains invalid characters,
// or does not yield any valid word parts.
var ErrUnsupported = errors.New("unsupported")

// Parsed holds the normalized lowercase word parts of a parsed identifier.
type Parsed struct {
	parts []string
}

// P parses s into a *Parsed instance, panicking via errorz.Check if parsing fails.
func P(s string) *Parsed {
	ret, err := Parse(s)
	errorz.Check(err)
	return ret
}

// MustParse parses s into a *Parsed instance, panicking if parsing fails. It is an alias for P.
func MustParse(s string) *Parsed {
	return P(s)
}

// Parse splits an input string into normalized lowercase word parts in a single pass.
//
// Splitting occurs at any hyphen ('-'), underscore ('_'), or uppercase Unicode letter.
// Consecutive or surrounding delimiters are ignored, and mixed conventions are supported.
//
// Examples:
//   - camelCase:   "fooBar"      -> ["foo", "bar"]
//   - PascalCase:  "FooBar"      -> ["foo", "bar"]
//   - Upper-run:   "ABC"         -> ["a", "b", "c"]
//   - snake_case:  "foo_bar"     -> ["foo", "bar"]
//   - kebab-case:  "foo-bar"     -> ["foo", "bar"]
//   - Mixed:       "a_b-c-pUi"   -> ["a", "b", "c", "p", "ui"]
//
// Returns ErrUnsupported if s contains invalid characters (spaces, symbols)
// or contains only delimiters without alphanumeric parts. An empty string returns
// a *Parsed with nil parts and no error.
func Parse(s string) (*Parsed, error) {
	if s == "" {
		return &Parsed{parts: nil}, nil
	}

	parts := make([]string, 0, 4)
	start := -1
	hasUpper := false

	for i, r := range s {
		switch {
		case r == '-' || r == '_':
			if start != -1 {
				part := s[start:i]
				if hasUpper {
					part = strings.ToLower(part)
				}
				parts = append(parts, part)
				start = -1
				hasUpper = false
			}
		case unicode.IsUpper(r):
			if start != -1 {
				part := s[start:i]
				if hasUpper {
					part = strings.ToLower(part)
				}
				parts = append(parts, part)
			}
			start = i
			hasUpper = true
		case unicode.IsLower(r) || unicode.IsDigit(r):
			if start == -1 {
				start = i
				hasUpper = false
			}
		default:
			return nil, fmt.Errorf("%w: invalid character in %q: %c", ErrUnsupported, s, r)
		}
	}

	if start != -1 {
		part := s[start:]
		if hasUpper {
			part = strings.ToLower(part)
		}
		parts = append(parts, part)
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("%w: no valid parts in %q", ErrUnsupported, s)
	}

	return &Parsed{parts: parts}, nil
}

// Parts returns a copy of the parsed lowercase word parts.
func (p *Parsed) Parts() []string {
	if p == nil || p.parts == nil {
		return nil
	}
	ret := make([]string, len(p.parts))
	copy(ret, p.parts)
	return ret
}

// Len returns the number of parts in the parsed identifier.
func (p *Parsed) Len() int {
	if p == nil {
		return 0
	}
	return len(p.parts)
}

// Get returns the part at index idx, or an empty string if p is nil or idx is out of bounds.
func (p *Parsed) Get(idx int) string {
	if p == nil || idx < 0 || idx >= len(p.parts) {
		return ""
	}
	return p.parts[idx]
}

// String returns the lower_snake_case representation of the parsed parts (equivalent to SL).
func (p *Parsed) String() string {
	return p.SL()
}

// C formats the parts into camelCase (e.g. "fooBar", "fooBarBaz").
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

// P formats the parts into PascalCase (e.g. "FooBar", "FooBarBaz").
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

// SU formats the parts into UPPER_SNAKE_CASE (e.g. "FOO_BAR").
func (p *Parsed) SU() string {
	if p == nil || len(p.parts) == 0 {
		return ""
	}
	return strings.ToUpper(strings.Join(p.parts, "_"))
}

// SL formats the parts into lower_snake_case (e.g. "foo_bar").
func (p *Parsed) SL() string {
	if p == nil || len(p.parts) == 0 {
		return ""
	}
	return strings.Join(p.parts, "_")
}

// QU formats the parts into UPPER-KEBAB-CASE (e.g. "FOO-BAR").
func (p *Parsed) QU() string {
	if p == nil || len(p.parts) == 0 {
		return ""
	}
	return strings.ToUpper(strings.Join(p.parts, "-"))
}

// QL formats the parts into lower-kebab-case (e.g. "foo-bar").
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
