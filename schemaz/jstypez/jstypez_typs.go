package jstypez

const (
	TypString  = "string"
	TypNumber  = "number"
	TypBoolean = "boolean"
	TypObject  = "object"
	TypAny     = "*"

	TypTime     = "dszTime"
	TypFileStat = "dszFileStat"
)

var typTime = Typedef{Head: &Tuple{Typ: "string", Name: TypTime, Summary: "Missing Summary"}}

var typFileStat = Typedef{
	Head: &Tuple{Typ: "object", Name: TypFileStat},
	Properties: NewTuples([]Tuple{
		{Typ: "string", Name: "url", Summary: "missing summary"},
		{Typ: "string", Name: "content_type", Summary: "missing summary"},
		{Typ: "string", Name: "md5", Summary: "missing summary"},
		{Typ: "string", Name: "size", Summary: "missing summary"},
		{Typ: "string", Name: "etag", Summary: "missing summary"},
		{Typ: TypTime, Name: "created_at", Summary: "missing summary"},
		{Typ: TypTime, Name: "updated_at", Summary: "missing summary"},
	}),
}

func BasicTypes() *Typedefs {
	var tds Typedefs
	tds.Add([]*Typedef{
		&typTime,
		&typFileStat,
	}...)
	return &tds
}
