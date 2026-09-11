package schemazv2

import (
	"github.com/infinity6-ai/gox/commonz/constraintz"
	"github.com/infinity6-ai/gox/commonz/strconvz"
)

type Parser[T any] struct {
	Parse  func(v T)
	Format func() T
}

func ParseStrNumber[O constraintz.Numbers](out *O) func() *Parser[string] {
	return func() *Parser[string] {
		return &Parser[string]{
			Parse: func(v string) {
				strconvz.MustParseNumberInto(v, out)
			},
			Format: func() string {
				return strconvz.FormatNumber(*out)
			},
		}
	}
}
