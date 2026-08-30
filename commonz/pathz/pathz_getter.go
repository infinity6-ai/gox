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
func (p *Path) Parent() (*Path, string) {
	// If the path has no parts (e.g., "", "."), or is the root ("/")
	if p.parts == "" {
		return nil, ""
	}

	lastSlashIndex := strings.LastIndex(p.parts, "/")
	if lastSlashIndex == -1 {
		// If there are no slashes, it's a single component path (e.g., "a", or "/a" where parts is "a").
		// In this case, there is no parent Path object.
		return nil, p.parts
	}

	parentParts := p.parts[:lastSlashIndex]
	base := p.parts[lastSlashIndex+1:]

	return &Path{parts: parentParts, parents: p.parents}, base
}

// Dir returns the parent Path, without the last element.
// For example, if the path is "a/b/c", it returns "a/b".
// If the path is "a", it returns "".
func (p *Path) Dir() *Path {
	parent, _ := p.Parent()
	return parent
}

func (p *Path) Base() string {
	_, ret := p.Parent()
	return ret
}
