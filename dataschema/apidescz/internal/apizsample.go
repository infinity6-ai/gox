package internal

import "github.com/infinity6-ai/gox/dataschema/schemaz"

type SampleApiParams struct {
	Numerator   float64
	Denominator float64
}

type SampleApiQuery struct {
	Precision int
}

type ReqMeta struct {
	TraceId string
}

type RespMeta struct {
	Hash string
}

type SampleApi struct {
	Name      string
	Method    string
	Path      string
	ReqParams SampleApiParams
	ReqQuery  SampleApiQuery
	ReqBody   schemaz.Field
	ReqMeta   ReqMeta
	RespMeta  RespMeta
	RespBody  schemaz.Field
}
