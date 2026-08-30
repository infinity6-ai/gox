package pathz

import "strings"

func (p *Path) IsEscaped() bool {
	return p.parents > 0
}

func (p *Path) IsAbsolute() bool {
	return p.parents == -1
}

func (p *Path) IsContained() bool {
	return p.parents == 0
}

func (p *Path) Split() []string {
	if p.parts == "" {
		return []string{}
	}
	return strings.Split(p.parts, "/")
}

// Parent returns the parent Path and the last element of the path.
// For example, if the path is "a/b/c", it returns ("a/b", "c").
// If the path is "a", it returns ("", "a").
func (p *Path) Parent() (parent *Path, base string, hasEndingSlash bool) {
	hasEndingSlash = p.hasEndingSlash
	if p.parts == "" {
		// Handles paths like "", ".", "/", "..", "../.."
		if p.parents > 0 {
			// This is an escaped path like "..", "../.."
			if p.parents == 1 {
				// For ".."
				return nil, "", hasEndingSlash
			}
			// For "../.." etc.
			return New(p.parents-1, []string{}, false), "", hasEndingSlash
		}
		// For "", "." or "/"
		return nil, "", hasEndingSlash
	}

	lastSlashIndex := strings.LastIndex(p.parts, "/")
	if lastSlashIndex == -1 {
		// Path has no slashes in its parts (e.g., "a", or "a" in "../a")
		// The parent is either just the '..' components, or nil for contained/absolute single part.
		if p.parents > 0 {
			// For "../a", parent is ".." represented as New(p.parents, []string{}, false)
			return New(p.parents, []string{}, false), p.parts, hasEndingSlash
		}
		// For "a" or "/a"
		return nil, p.parts, hasEndingSlash
	}

	parentPartsStr := p.parts[:lastSlashIndex]
	base = p.parts[lastSlashIndex+1:]

	// Reconstruct the parent Path using New to handle internal structure correctly
	// Split parentPartsStr to pass to New constructor
	parentPartsSlice := strings.Split(parentPartsStr, "/")
	return New(p.parents, parentPartsSlice, false), base, hasEndingSlash
}

// Dir returns the parent Path, without the last element.
// For example, if the path is "a/b/c", it returns "a/b".
// If the path is "a", it returns "".
func (p *Path) Dir() *Path {
	parent, _, _ := p.Parent()
	return parent
}

func (p *Path) Base() string {
	_, ret, _ := p.Parent()
	return ret
}

func (p *Path) BaseSlash() (base string, hasEndingSlash bool) {
	_, base, hasEndingSlash = p.Parent()
	return
}
