package pathz

import (
	"github.com/infinity6-ai/gox/commonz/validation"
)

type ValidateOptions struct {
	Parents     *int
	Wildchar    bool
	EndingSlash *bool
}

func (p *Path) Validate(opts ValidateOptions) error {
	if opts.Parents != nil {
		err := validation.Equal(*opts.Parents, p.parents, "path parents mismatch for path %s", p)
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
