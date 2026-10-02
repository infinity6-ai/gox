package schemaz

import (
	"strconv"

	"github.com/infinity6-ai/gox/commonz/constraintz"
	"github.com/infinity6-ai/gox/commonz/strconvz"
)

type Parser[T any] struct {
	Parse  func(v T)
	Format func() (T, any)
}

func ParseStrNumber[O constraintz.Numbers](out *O) func() *Parser[string] {
	return func() *Parser[string] {
		return &Parser[string]{
			Parse: func(v string) {
				strconvz.MustParseNumberInto(v, out)
			},
			Format: func() (string, any) {
				return strconvz.FormatNumber(*out), *out
			},
		}
	}
}

func ParseStr(out *string) func() *Parser[string] {
	return func() *Parser[string] {
		return &Parser[string]{
			Parse: func(v string) {
				*out = v
			},
			Format: func() (string, any) {
				return *out, *out
			},
		}
	}
}

func ParseStrBool(out *bool) func() *Parser[string] {
	return func() *Parser[string] {
		return &Parser[string]{
			Parse: func(v string) {
				*out = strconvz.MustParseBool(v)
			},
			Format: func() (string, any) {
				return strconv.FormatBool(*out), *out
			},
		}
	}
}

func ParseStrGet[T any](out *T, set func(v string) T) func() *Parser[string] {
	return func() *Parser[string] {
		return &Parser[string]{
			Parse: func(v string) {
				*out = set(v)
			},
			Format: func() (string, any) {
				type G interface {
					Get() string
				}
				return any(out).(G).Get(), *out
			},
		}
	}
}
