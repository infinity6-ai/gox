package pathz

import (
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/validation"
)

type ValidateOptions struct {
	Parents     *int
	Wildchar    bool
	EndingSlash *bool
}

// Validate checks the path against the provided options and panics if a validation fails.
// - `Parents`: If not nil, it checks if the number of parent directory traversals (`..`) matches the specified value.
// - `EndingSlash`: If not nil, it checks if the path's trailing slash status matches the specified boolean.
// - `Wildchar`: If false, it panics if any part of the path contains wildcard characters ('{' or '}').
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
			if strings.Contains(part, "*") {
				panic(fmt.Errorf("path contains wildcard characters when not allowed: %s", p))
			}
		}
	}
}
