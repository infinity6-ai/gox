package pathz

import (
	"github.com/infinity6-ai/gox/commonz/validation"
)

type ValidateOptions struct {
	Absolute    *bool
	MaxParents  *int
	Wildchar    bool
	EndingSlash *bool
	Empty       *bool
}

func (p *Path) Validate(opts ValidateOptions) error {
	if opts.Absolute != nil {
		err := validation.Equal(*opts.Absolute, p.IsAbsolute(), "path absolute flag: %s", p)
		if err != nil {
			return err
		}
	} else if opts.MaxParents != nil {
		err := validation.GreaterOrEqual(p.Parents(), *opts.MaxParents, "max parents allowed: %s", p)
		if err != nil {
			return err
		}
	}
	if opts.Empty != nil {
		err := validation.Equal(*opts.Empty, len(p.Parts()) == 0, "path empty flag: %s", p)
		if err != nil {
			return err
		}
	}
	if opts.EndingSlash != nil {
		err := validation.Equal(*opts.EndingSlash, p.hasEndingSlash, "path ending slash mismatch for path %s", p)
		if err != nil {
			return err
		}
	}
	if !opts.Wildchar {
		for _, part := range p.parts {
			err := validation.StrContains("*", part, "path contains wildcard characters when not allowed")
			if err != nil {
				return err
			}
		}
	}
	return nil
}
