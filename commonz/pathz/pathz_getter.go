package pathz

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
	if p.parts == nil {
		return []string{}
	}
	return p.parts
}

// Parent returns the parent Path and the last element of the path.
// For example, if the path is "a/b/c", it returns ("a/b", "c").
// If the path is "a", it returns ("", "a").
func (p *Path) Parent() (parent *Path, base string, hasEndingSlash bool) {
	hasEndingSlash = p.hasEndingSlash
	if len(p.parts) == 0 {
		// Handles paths like "", ".", "/", "..", "../.."
		if p.parents > 0 {
			// This is an escaped path like "..", "../.."
			if p.parents == 1 {
				// For ".."
				return nil, "", false
			}
			// For "../.." etc.
			return New(p.parents-1, []string{}, false), "", hasEndingSlash
		}
		// For "", "." or "/"
		return nil, "", false
	}

	base = p.parts[len(p.parts)-1]
	if len(p.parts) == 1 {
		// Path has only one part (e.g., "a", or "a" in "../a")
		// The parent is either just the '..' components, or nil for contained/absolute single part.
		if p.parents > 0 {
			// For "../a", parent is ".." represented as New(p.parents, []string{}, false)
			return New(p.parents, []string{}, false), base, hasEndingSlash
		}
		// For "a" or "/a"
		return nil, base, false
	}

	parentPartsSlice := p.parts[:len(p.parts)-1]
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
