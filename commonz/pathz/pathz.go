package pathz

type Part struct {
	Name        string
	Placeholder bool
}

type Patttern struct {
	Parts []*Part
}

type Path struct {
	Pattern *Patttern
	Values  map[string]string
}

func (p *Path) String() string {
	panic("implement it")
}

func Parse(path string) *Path {
	panic("implement it")
}
