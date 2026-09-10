package schemazv2

type Desc struct {
	Name     string
	Summary  string
	Markdown string
}

type Schema struct {
	Raw    func() any
	Object func() map[string]*Schema
	Array  func() *Schema
	Desc   func() *Desc
}
