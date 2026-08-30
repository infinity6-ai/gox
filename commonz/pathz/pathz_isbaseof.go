package pathz

func (p *Path) IsBaseOf(other *Path) bool {
	if p.parents != other.parents {
		return false
	}
	for idx, part := range p.parts {
		if idx >= len(other.parts) {
			return false
		}
		if part != other.parts[idx] {
			return false
		}
	}
	if len(other.parts) > len(p.parts) {
		return true
	}
	if !p.hasEndingSlash {
		return true
	}
	return other.hasEndingSlash
}
