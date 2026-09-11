package schemazv2

import (
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
