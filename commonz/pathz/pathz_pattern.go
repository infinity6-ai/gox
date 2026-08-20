package pathz

type Part struct {
	Name        string
	Placeholder bool
}

type Pattern struct {
	Parts []*Part
}
