package schemaz

type Type string

const (
	TypeObject  Type = "object"
	TypeArray   Type = "array"
	TypeString  Type = "string"
	TypeNumber  Type = "number"
	TypeBoolean Type = "boolean"
)

type Desc struct {
	Name     string
	Summary  string
	Markdown string
}

type Spec struct {
	Type Type `json:"type"`

	// array
	ArrayType Type `json:"array_type"`

	// object
	Fields []Field `json:"fields"`
}

type Field struct {
	Name string `json:"name"`
	Desc Desc   `json:"desc"`
	Spec Spec   `json:"spec"`
}

type Api[I any, O any] struct {
	Id         string
	Desc       Desc
	Method     string
	Path       string
	ReqParams  []Field // basics
	ReqQuery   []Field // basics
	ReqHeaders []Field // basics

	ReqBody *Spec
	ReqType I

	RespHeaders []Field // basics

	RespBody *Spec
	RespType O
}

func (a Api[I, O]) NewReqBody() I {
	var b I
	return b
}

func (a Api[I, O]) NewRespBody() O {
	var b O
	return b
}
