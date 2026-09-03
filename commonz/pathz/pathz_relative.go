package pathz

import (
	"fmt"
)

// ExtractRelative calculates the relative path from a base path `p` to another path `other`.
// It returns a new Path object representing the relative path if `other` is a descendant of `p`.
// If `other` is not a descendant of `p`, it returns an `ErrNavigationError`.
func (p *Path) ExtractRelative(other *Path) (*Path, error) {
	if !p.IsBaseOf(other) {
		return nil, fmt.Errorf("%w: '%s' is not relative to '%s'", ErrNavigationError, other, p)
	}

	// If paths are identical, the relative path is "."
	if len(p.parts) == len(other.parts) {
		return New(0, nil, false), nil
	}

	relativeParts := other.parts[len(p.parts):]
	return New(0, relativeParts, other.hasEndingSlash), nil
}

func (p *Path) ForceCurrent() *Path {
	ret := p.Clone()
	ret.parents = 0
	return ret
}
