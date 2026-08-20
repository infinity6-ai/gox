package pathz

type Part struct {
	name        string
	placeholder bool
}

type Patttern struct {
	parts []*Part
}

type Path struct {
	pattern *Patttern
	values  map[string]string
}
