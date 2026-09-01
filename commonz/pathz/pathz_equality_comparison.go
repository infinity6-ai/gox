package pathz

import "cmp"

// Equals checks if two Path objects are equal.
// Two paths are considered equal if all their components (parents, parts, hasEndingSlash) are identical.
func (p *Path) Equals(other *Path) bool {
	if p == nil && other == nil {
		return true
	}
	if p == nil || other == nil {
		return false
	}
	if p.parents != other.parents {
		return false
	}
	if p.hasEndingSlash != other.hasEndingSlash {
		return false
	}
	if len(p.parts) != len(other.parts) {
		return false
	}
	for i := range p.parts {
		if p.parts[i] != other.parts[i] {
			return false
		}
	}
	return true
}

// Compare compares two Path objects based on their string representation.
// It returns:
// - 0 if the paths are equal (p.String() == other.String())
// - -1 if p's string representation is lexicographically smaller than other's
// - 1 if p's string representation is lexicographically greater than other's
func (p *Path) Compare(other *Path) int {
	return cmp.Compare(p.String(), other.String())
}
