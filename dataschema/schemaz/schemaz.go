package schemaz

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
	ArrayType *Field `json:"array_type"`

	// object
	Fields map[string]*Field `json:"fields"`
}
