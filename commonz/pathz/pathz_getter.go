package pathz

func (p *Path) IsEscaped() bool {
	return p.parents != 0
}

func (p *Path) IsAbsolute() bool {
	return p.parents == -1
}

func (p *Path) IsContained() bool {
	return p.parents == 0
}
