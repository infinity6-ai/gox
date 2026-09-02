package dataschema

type Type string

const (
	TypeObject  Type = "object"
	TypeArray   Type = "array"
	TypeString  Type = "string"
	TypeNumber  Type = "number"
	TypeBoolean Type = "boolean"
)

type Field struct {
	Type Type `json:"type"`

	// array
	Field *Field `json:"field"`

	// object
	Fields map[string]*Field `json:"fields"`
}

func (f *Field) Validate(name string, t Type) error {
	return nil
}

type Schema struct {
	Name string `json:"name"`
	Root *Field `json:"root"`
}
