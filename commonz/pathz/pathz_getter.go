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

func (p *Path) Base() string {
	if p.parts == "" {
		return ""
	}
	lastIndex := strings.LastIndex(p.parts, "/")
	if lastIndex == -1 {
		return p.parts
	}
	return p.parts[lastIndex+1:]
}

// Parent returns the parent Path and the last element of the path.
// For example, if the path is "a/b/c", it returns ("a/b", "c").
// If the path is "a", it returns ("", "a").
func (p *Path) Parent() (*Path, string) {
	if p.parts == "" {
		return &Path{parts: "", parents: p.parents}, ""
	}

	lastSlashIndex := strings.LastIndex(p.parts, "/")
	if lastSlashIndex == -1 { // No slash in parts
		if p.IsAbsolute() {
			// for path "/a", parent is "/"
			return &Path{parts: "", parents: p.parents}, p.parts
		}
		// for path "a", parent is nil
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
	if parent == nil {
		return &Path{parts: "", parents: p.parents}
	}
	return parent
}
