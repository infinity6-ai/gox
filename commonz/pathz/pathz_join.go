package pathz

import (
	"errors"
	"fmt"
)

var ErrEscaped = errors.New("path escaped error")

// Join combines the receiver path `p` with one or more other paths and returns the resulting path.
// It intelligently handles path separators, parent directory traversals (`..`), and absolute paths.
// If any of the `others` paths is absolute, it discards the preceding paths (including the receiver `p`)
// and builds from that absolute path.
//
// A critical feature of this method is its safety check: it ensures that the final, resolved path
// does not "escape" or navigate outside the hierarchy of the original base path `p`. If the resulting
// path is not a descendant of or equal to `p`, the function will return the escaped `*Path` and
// an `ErrEscaped` error.
//
// For example:
//   - `p("a/b").Join(p("c"))` results in `p("a/b/c")`.
//   - `p("a/b").Join(p("../c"))` results in `p("a/c")`.
//   - `p("a/b").Join(p("../../c"))` returns `ErrEscaped` because it navigates outside `a/b`.
//   - `p("a/b").Join(p("/c"))` returns `ErrEscaped` because `/c` is not relative to `a/b`.
func (p *Path) Join(others ...*Path) (*Path, error) {
	// if len(others) == 0 {
	// 	return p, nil
	// }

	ret := p.Clone()

	for _, other := range others {
		other = other.Clone()
		if other.IsAbsolute() {
			ret = other
			continue
		}
		// ../a + ../../b => ./b
		step := len(p.parts) - other.parents
		if step < 0 {
			// 2 - 1 = 1
			other.parents = other.parents - len(p.parts)
			// 1
			step = len(p.parts)
		}
		if step > 0 {
			// ["a"] 0 -> 1 - 1 = 0 = []
			ret.parts = other.parts[:len(other.parts)-step]
		}
		// 1 + 1 = 2
		ret.parents = ret.parents + other.parents
		ret.hasEndingSlash = other.hasEndingSlash
	}

	if !p.IsBaseOf(ret) {
		return ret, fmt.Errorf("%w: joining '%s' to '%s' results in '%s' which is outside the base", ErrEscaped, p, others, ret)
	}
	return ret, nil

	// otherStrs := make([]string, len(others))
	// for i, o := range others {
	// 	otherStrs[i] = o.String()
	// }

	// resultPathStr := path.Join(append([]string{p.String()}, otherStrs...)...)

	// resultPath, err := Parse(resultPathStr)
	// if err != nil {
	// 	return nil, fmt.Errorf("cannot parse joined path: %w", err)
	// }

	// if !p.IsBaseOf(resultPath) {
	// 	return resultPath, fmt.Errorf("%w: joining '%s' to '%s' results in '%s' which is outside the base", ErrEscaped, strings.Join(otherStrs, "/"), p, resultPath)
	// }

	// return resultPath, nil
}
