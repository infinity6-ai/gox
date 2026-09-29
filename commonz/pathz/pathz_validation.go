package pathz

import (
	"fmt"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/validation"
)

type ValidateOptions struct {
	Absolute    *bool
	MaxParents  *int
	Wildchar    bool
	EndingSlash *bool
	MinPart     int
	MaxPart     *int
	Part        func(idx int, part string) error
}

func (p *Path) Check(opts ValidateOptions) {
	errorz.Check(p.Validate(opts))
}

func (p *Path) Validate(opts ValidateOptions) error {
	if opts.Absolute != nil {
		err := validation.Equal(*opts.Absolute, p.IsAbsolute(), "path absolute flag: %s", p)
		if err != nil {
			return err
		}
	}
	if opts.MaxParents != nil {
		if *opts.MaxParents < 0 {
			panic("max parents cannot be negative, use opts.Absolute instead")
		}
		if p.IsAbsolute() {
			return fmt.Errorf("max parents allowed %d, but is is absolute: %s", *opts.MaxParents, p)
		}
		if p.Parents() < 0 {
			panic("parents cannot be lesser than -1")
		}
		err := validation.LessOrEqual(p.Parents(), *opts.MaxParents, "max parents allowed: %s", p)
		if err != nil {
			return err
		}
	}
	if opts.MinPart < 0 {
		panic("min part cannot be negative")
	}
	if opts.MaxPart != nil {
		if *opts.MaxPart < 0 {
			panic("max part cannot be negative")
		}
		if opts.MinPart > *opts.MaxPart {
			panic("min part cannot be greater than max part")
		}
		err := validation.LessOrEqual(len(p.parts), *opts.MaxPart, "max parts allowed: %s", p)
		if err != nil {
			return err
		}
	}
	if opts.MinPart > 0 {
		err := validation.GreaterOrEqual(len(p.parts), opts.MinPart, "min parts allowed: %s", p)
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
			err := validation.StrNotContains("*", part, "path contains wildcard characters when not allowed")
			if err != nil {
				return err
			}
		}
	}
	if opts.Part != nil {
		for idx, part := range p.parts {
			err := opts.Part(idx, part)
			if err != nil {
				return fmt.Errorf("part %d (%s) validation failed: %w", idx, part, err)
			}
		}
	}
	return nil
}

func (p *Path) ValidateAbsoluteFile() error {
	return p.Validate(ValidateOptions{
		Absolute: new(true),
		Wildchar: false,
	})
}
