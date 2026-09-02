package apidescz

type Api[I any, O any] struct {
	Name   string
	Groups []string
	Short  string
	Guide  string
	Path   string
	Input  I
	Output O
}

type Collection struct {
}
