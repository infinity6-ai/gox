package pathz

import (
	"fmt"
	"path"
	"strings"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

// Join concatenates a Path with one or more other Paths.
//
// The `result` Path is the concatenation of the receiver Path and all `others` Paths.
// The `contained` Path is identical to `result` if `result` is a contained path (i.e., not absolute and not escaped).
// Otherwise, `contained` is nil.
func (p *Path) Join(others ...*Path) (contained *Path, result *Path) {
	// c := p.String()
	// for _, o := range others {
	// 	c = path.Join(c, o.String())
	// }
	// ret, err := Parse(c)
	// errorz.Check(err)
	// if !p.IsBaseOf(ret) {
	// 	return nil, ret
	// }
	// return ret, ret

	// Start with the receiver path's string representation and ending slash status.
	// We will manually build the joined string to account for the unexpected behavior of path.Join
	// regarding absolute path overriding in this environment.

	currentString := p.String()
	currentHasEndingSlash := p.hasEndingSlash

	// Iterate through the other paths to join.
	for _, other := range others {
		if other == nil {
			continue // Skip nil paths.
		}

		otherString := other.String()

		// If the 'other' path is absolute, it overrides all previous parts.
		// We detect this by checking if its string representation starts with "/".
		if strings.HasPrefix(otherString, "/") {
			currentString = otherString
			currentHasEndingSlash = other.hasEndingSlash
			continue // Start new path from this absolute path.
		}

		// If currentString is empty (e.g., from an initial empty path or after an absolute path override)
		// then we just take the otherString.
		if currentString == "" {
			currentString = otherString
			currentHasEndingSlash = other.hasEndingSlash
		} else {
			// Otherwise, concatenate with a slash. Use path.Join here to handle cleaning of ".." and "." components.
			currentString = path.Join(currentString, otherString)
			currentHasEndingSlash = other.hasEndingSlash
		}
	}

	// After manual concatenation, ensure the final path string correctly reflects the ending slash.
	// The path.Join above might strip a trailing slash if the 'currentString' had one but 'otherString' didn't.
	// We need to re-add it if the last contributing path had one and it's not already there.
	if currentHasEndingSlash && currentString != "/" && currentString != "" && !strings.HasSuffix(currentString, "/") {
		currentString += "/"
	}

	// Parse the final manually constructed string into a Path object.
	newPath, err := Parse(currentString)
	if err != nil {
		// This indicates an internal error if a valid string cannot be parsed.
		errorz.Check(fmt.Errorf("pathz.Join: internal error parsing manually joined path '%s': %w", currentString, err))
	}
	result = newPath

	// Determine the 'contained' result.
	if result.IsContained() {
		contained = result
	} else {
		contained = nil
	}

	return contained, result
}
